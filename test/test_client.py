"""The client sends, streams, and watches cursor positions through the server."""

import queue
import subprocess
import threading
import time

import harness

USERNAME = "demo_user"
OTHER = "other_user"
PASSWORD = harness.PASSWORD
SEND, TRACK, SHARE = 1, 2, 3
LINE_TIMEOUT = 10.0


class Lines:
    """The lines a process prints, read in the background so a test can wait for each in turn."""

    def __init__(self, process: subprocess.Popen[str]) -> None:
        self.lines: queue.Queue[str] = queue.Queue()
        stdout = process.stdout
        assert stdout is not None
        threading.Thread(
            target=lambda: [self.lines.put(line) for line in stdout], daemon=True
        ).start()

    def next(self) -> str:
        """Return the next line, failing if none arrives in time."""
        try:
            return self.lines.get(timeout=LINE_TIMEOUT)
        except queue.Empty:
            raise TimeoutError(f"no output within {LINE_TIMEOUT}s") from None


class ClientTest(harness.IntegrationTestCase):
    def setUp(self) -> None:
        super().setUp()
        self.server.signup(USERNAME)

    def test_sends_a_position(self) -> None:
        first = self.run_client(USERNAME, PASSWORD, "send", "0.25", "0.75")
        second = self.run_client(USERNAME, PASSWORD, "send", "1", "0")

        self.assertEqual(first.returncode, 0, first.stderr)
        self.assertEqual(second.returncode, 0, second.stderr)
        # Each Send is a stream of its own.
        self.assertEqual(first.stdout, "1\n")
        self.assertEqual(second.stdout, "2\n")
        self.assertEqual(
            self.server.cursors(), [(USERNAME, SEND, 0.25, 0.75), (USERNAME, SEND, 1.0, 0.0)]
        )
        # The client logs out after sending, so its session does not outlive the run.
        self.assertEqual(self.server.sessions(), 0)

    def test_username_is_case_insensitive(self) -> None:
        result = self.run_client(USERNAME.upper(), PASSWORD, "send", "0.5", "0.5")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.server.cursors(), [(USERNAME, SEND, 0.5, 0.5)])

    def test_rejects_position_out_of_range(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "send", "1.5", "0.5")

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.cursors(), [])

    def test_tracks_piped_positions(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "track", input="0 0\n\n0.3 0.4\n0.3 0\n")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "1 3 0.9000\n")
        self.assertEqual(
            self.server.cursors(),
            [(USERNAME, TRACK, 0.0, 0.0), (USERNAME, TRACK, 0.3, 0.4), (USERNAME, TRACK, 0.3, 0.0)],
        )
        self.assertEqual(self.server.sessions(), 0)

    def test_tracks_nothing_without_positions(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "track", input="")

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "0 0 0.0000\n")
        self.assertEqual(self.server.cursors(), [])

    def test_watches_positions(self) -> None:
        watch = Lines(self.start_client(USERNAME, PASSWORD, "watch"))
        # The watcher has subscribed once it logged in, which shows as its session.
        deadline = time.monotonic() + LINE_TIMEOUT
        while self.server.sessions() == 0:
            self.assertLess(time.monotonic(), deadline, "the watcher never logged in")
            time.sleep(0.05)
        time.sleep(0.5)

        tracked = self.run_client(USERNAME, PASSWORD, "track", input="0.1 0.2\n")
        self.assertEqual(tracked.returncode, 0, tracked.stderr)

        self.assertEqual(watch.next(), f"moved 1 {USERNAME} track 0.1000 0.2000\n")
        self.assertEqual(watch.next(), f"ended 1 {USERNAME}\n")

    def test_shares_positions(self) -> None:
        self.server.signup(OTHER)
        other = self.start_client(OTHER, PASSWORD, "share")
        other_lines = Lines(other)
        deadline = time.monotonic() + LINE_TIMEOUT
        while self.server.sessions() == 0:
            self.assertLess(time.monotonic(), deadline, "the sharer never logged in")
            time.sleep(0.05)
        time.sleep(0.5)

        # The first sharer hears the second's position, and the second hears the first's.
        shared = self.start_client(USERNAME, PASSWORD, "share")
        shared_lines = Lines(shared)
        assert shared.stdin is not None and other.stdin is not None
        time.sleep(0.5)
        shared.stdin.write("0.5 0.25\n")
        shared.stdin.flush()
        self.assertEqual(other_lines.next(), f"moved 1 {USERNAME} share 0.5000 0.2500\n")
        other.stdin.write("0.75 1\n")
        other.stdin.flush()
        self.assertEqual(shared_lines.next(), f"moved 2 {OTHER} share 0.7500 1.0000\n")

        # Ending one sharer's input ends its call, and the other hears its stream end.
        shared.stdin.close()
        self.assertEqual(shared.wait(timeout=LINE_TIMEOUT), 0)
        self.assertEqual(other_lines.next(), f"ended 1 {USERNAME}\n")
        self.assertEqual(
            self.server.cursors(), [(USERNAME, SHARE, 0.5, 0.25), (OTHER, SHARE, 0.75, 1.0)]
        )


if __name__ == "__main__":
    harness.main()
