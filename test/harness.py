"""Shared helpers for integration tests that drive the built client, server, and web app.

Each test script runs as `python3 test_*.py <client> <server>`, where the
arguments are paths to the client and server executables.
"""

import argparse
import contextlib
import os
import signal
import socket
import sqlite3
import ssl
import subprocess
import sys
import tempfile
import time
import unittest
from dataclasses import dataclass
from pathlib import Path

from playwright.sync_api import Browser, Page, Playwright, sync_playwright

PASSWORD = "example-password"
# A bcrypt hash of PASSWORD, so users can be created without the API.
PASSWORD_HASH = "$2y$10$L3wFWn4QMErfVV9g.bhqbu1NmubtoyVhDiK2mySGQaft2RSooXRQS"
READY_TIMEOUT = 10.0
STOP_TIMEOUT = 10.0
CLIENT_TIMEOUT = 30.0
BROWSER_TIMEOUT = 10_000  # milliseconds


@dataclass(frozen=True)
class Binaries:
    client: Path
    server: Path


_binaries: Binaries | None = None


def binaries() -> Binaries:
    if _binaries is None:
        raise RuntimeError("harness.main() was not called")
    return _binaries


def main() -> None:
    """Parse the executable paths, then run the calling module's tests."""
    global _binaries

    parser = argparse.ArgumentParser()
    parser.add_argument("client", type=Path, help="path to the client executable")
    parser.add_argument("server", type=Path, help="path to the server executable")
    args, rest = parser.parse_known_args()
    _binaries = Binaries(client=args.client.absolute(), server=args.server.absolute())

    unittest.main(module="__main__", argv=[sys.argv[0], *rest])


def free_port() -> int:
    """Return a port that is free for both TCP and UDP, which the server listens on together."""
    while True:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as tcp:
            tcp.bind(("127.0.0.1", 0))
            port = tcp.getsockname()[1]
            with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as udp:
                try:
                    udp.bind(("127.0.0.1", port))
                except OSError:
                    continue
            return port


class Server:
    """A server process with its own home directory, and therefore its own database.

    The server generates a self-signed certificate, which the browser trusts for
    WebTransport by the hash the server publishes, and the client by its file.
    """

    def __init__(self, binary: Path, home: Path) -> None:
        self.home = home
        self.port = free_port()
        self.process = subprocess.Popen(
            [binary, "--port", str(self.port)],
            env={
                "PATH": os.environ.get("PATH", ""),
                "HOME": str(home),
                "XDG_CONFIG_HOME": str(home / ".config"),
            },
        )

    @property
    def origin(self) -> str:
        return f"https://127.0.0.1:{self.port}"

    @property
    def certificate(self) -> Path:
        return self.home / "certificate.pem"

    @property
    def database(self) -> Path:
        # The server stores its database under os.UserConfigDir(), which differs by platform.
        return next(self.home.rglob("trevstack.db"))

    def wait_ready(self) -> None:
        deadline = time.monotonic() + READY_TIMEOUT
        while time.monotonic() < deadline:
            if (code := self.process.poll()) is not None:
                raise RuntimeError(f"server exited with code {code} before it was ready")
            with contextlib.suppress(OSError):
                certificate = ssl.get_server_certificate(("127.0.0.1", self.port), timeout=1)
                self.certificate.write_text(certificate)
                return
            time.sleep(0.05)
        raise TimeoutError(f"server was not ready within {READY_TIMEOUT}s")

    def stop(self) -> None:
        if self.process.poll() is not None:
            return
        self.process.send_signal(signal.SIGINT)
        try:
            self.process.wait(timeout=STOP_TIMEOUT)
        except subprocess.TimeoutExpired:
            self.process.kill()
            self.process.wait()

    def signup(self, username: str) -> None:
        """Create a user with PASSWORD directly in the database."""
        with contextlib.closing(sqlite3.connect(self.database)) as db, db:
            db.execute(
                "INSERT INTO users (username, password_hash) VALUES (?, ?)",
                (username.lower(), PASSWORD_HASH),
            )

    def numbers(self) -> list[tuple[str, int]]:
        with contextlib.closing(sqlite3.connect(self.database)) as db:
            return db.execute("SELECT name, number FROM numbers ORDER BY id").fetchall()

    def sessions(self) -> int:
        """Return how many sessions the server keeps, expired or not."""
        with contextlib.closing(sqlite3.connect(self.database)) as db:
            return db.execute("SELECT count(*) FROM sessions").fetchone()[0]


class IntegrationTestCase(unittest.TestCase):
    """Starts a fresh server with an empty database for every test."""

    server: Server

    def setUp(self) -> None:
        home = tempfile.TemporaryDirectory()
        self.addCleanup(home.cleanup)
        self.server = Server(binaries().server, Path(home.name))
        self.addCleanup(self.server.stop)
        self.server.wait_ready()

    def run_client(
        self, username: str, password: str, name: str, number: int
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [binaries().client, name, str(number)],
            env={
                "PATH": os.environ.get("PATH", ""),
                "STACK_URL": self.server.origin,
                "STACK_CA_CERT": str(self.server.certificate),
                "STACK_USERNAME": username,
                "STACK_PASSWORD": password,
            },
            capture_output=True,
            text=True,
            timeout=CLIENT_TIMEOUT,
        )


class BrowserTestCase(IntegrationTestCase):
    """Also opens a page on the server's web app, in a fresh browser context for every test."""

    playwright: Playwright
    browser: Browser
    page: Page

    @classmethod
    def setUpClass(cls) -> None:
        super().setUpClass()
        cls.playwright = sync_playwright().start()
        cls.addClassCleanup(cls.playwright.stop)
        cls.browser = cls.playwright.chromium.launch()
        cls.addClassCleanup(cls.browser.close)

    def setUp(self) -> None:
        super().setUp()
        # A service worker would serve cached assets and hide what the server returns.
        # WebTransport ignores ignore_https_errors and trusts the certificate by its hash.
        context = self.browser.new_context(
            base_url=self.server.origin, service_workers="block", ignore_https_errors=True
        )
        self.addCleanup(context.close)
        context.set_default_timeout(BROWSER_TIMEOUT)
        self.page = context.new_page()
