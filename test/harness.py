"""Shared helpers for integration tests that drive the built client and server.

Each test script runs as `python3 test_*.py <client> <server>`, where the
arguments are paths to the client and server executables.
"""

import argparse
import contextlib
import json
import os
import signal
import socket
import sqlite3
import subprocess
import sys
import tempfile
import time
import unittest
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any

JWT_SECRET = "integration-test-secret-that-is-long-enough"
READY_TIMEOUT = 10.0
STOP_TIMEOUT = 10.0
CLIENT_TIMEOUT = 30.0


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
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


class Server:
    """A server process with its own home directory, and therefore its own database."""

    def __init__(self, binary: Path, home: Path) -> None:
        self.home = home
        self.port = free_port()
        self.process = subprocess.Popen(
            [binary, "--port", str(self.port)],
            env={
                "PATH": os.environ.get("PATH", ""),
                "HOME": str(home),
                "XDG_CONFIG_HOME": str(home / ".config"),
                "JWT_SECRET": JWT_SECRET,
            },
        )

    @property
    def url(self) -> str:
        return f"http://127.0.0.1:{self.port}/grpc"

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
                socket.create_connection(("127.0.0.1", self.port), timeout=1).close()
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

    def call(self, procedure: str, message: dict[str, Any]) -> Any:
        """Call a unary RPC using the Connect protocol with JSON encoding."""
        request = urllib.request.Request(
            f"{self.url}/{procedure}",
            data=json.dumps(message).encode(),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(request, timeout=10) as response:
            return json.load(response)

    def signup(self, username: str, password: str) -> None:
        self.call("auth.v1.AuthService/Signup", {"username": username, "password": password})

    def numbers(self) -> list[tuple[str, int]]:
        with contextlib.closing(sqlite3.connect(self.database)) as db:
            return db.execute("SELECT name, number FROM numbers ORDER BY id").fetchall()


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
                "STACK_URL": self.server.url,
                "STACK_USERNAME": username,
                "STACK_PASSWORD": password,
            },
            capture_output=True,
            text=True,
            timeout=CLIENT_TIMEOUT,
        )
