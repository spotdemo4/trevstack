"""The web app signs users up and in, sends and shows cursor positions, and signs users out."""

import re
import time
from collections.abc import Callable

import harness
from playwright.sync_api import expect

USERNAME = "demo_user"
PASSWORD = harness.PASSWORD
# cursor.v1.Source of a Send call
SOURCE_SEND = 1


class WebTest(harness.BrowserTestCase):
    def login(self, path: str) -> None:
        """Sign in as USERNAME, who must exist, from a protected path."""
        self.page.goto(path)
        login = self.page.get_by_role("tabpanel", name="Login")
        login.get_by_label("Username").fill(USERNAME)
        login.get_by_label("Password").fill(PASSWORD)
        login.get_by_role("button", name="Login").click()
        expect(self.page).to_have_url(re.compile(rf"{re.escape(path)}$"))

    def wait_for(self, condition: Callable[[], bool], message: str) -> None:
        deadline = time.monotonic() + harness.BROWSER_TIMEOUT / 1000
        while not condition() and time.monotonic() < deadline:
            time.sleep(0.05)
        self.assertTrue(condition(), message)

    def test_redirects_to_sign_in(self) -> None:
        self.page.goto("/tables")

        expect(self.page).to_have_url(re.compile(r"/auth\?"))
        expect(self.page.get_by_role("heading", name="Welcome to TrevStack")).to_be_visible()

    def test_signup_then_list_positions(self) -> None:
        self.page.goto("/tables")
        self.page.get_by_role("tab", name="Sign up").click()
        signup = self.page.get_by_role("tabpanel", name="Sign up")
        signup.get_by_label("Username").fill(USERNAME)
        signup.get_by_label("Password").fill(PASSWORD)
        signup.get_by_role("button", name="Create account").click()

        expect(self.page).to_have_url(re.compile(r"/tables$"))
        expect(self.page.get_by_text("No items found.")).to_be_visible()

        result = self.run_client(USERNAME, PASSWORD, "send", "0.25", "0.75")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.page.reload()

        positions = self.page.get_by_role("tabpanel", name="Positions")
        row = positions.get_by_role("row").filter(has_text=USERNAME)
        expect(row).to_contain_text("Send")
        expect(row).to_contain_text("0.250")
        expect(row).to_contain_text("0.750")

        self.page.get_by_role("tab", name="Streams").click()
        streams = self.page.get_by_role("tabpanel", name="Streams")
        expect(streams.get_by_role("row").filter(has_text=USERNAME)).to_contain_text("Send")

    def test_forms_sends_the_position_picked_on_the_pad(self) -> None:
        self.server.signup(USERNAME)
        self.login("/forms")

        pad = self.page.get_by_label("Position pad")
        box = pad.bounding_box()
        assert box is not None
        pad.click(position={"x": box["width"] * 0.25, "y": box["height"] * 0.75})
        expect(self.page.get_by_label("X", exact=True)).to_have_value("0.25")
        expect(self.page.get_by_label("Y", exact=True)).to_have_value("0.75")

        self.page.get_by_role("button", name="Send").click()

        expect(self.page.get_by_text("Position sent")).to_be_visible()
        self.wait_for(
            lambda: self.server.cursors() == [(USERNAME, SOURCE_SEND, 0.25, 0.75)],
            f"unexpected positions: {self.server.cursors()}",
        )

    def test_server_streams_shows_positions_live(self) -> None:
        self.server.signup(USERNAME)
        self.login("/server-streams")

        # The page may not have started watching yet, so send until the cursor shows.
        cursor = self.page.get_by_label("Live cursors").get_by_text(USERNAME)
        for _ in range(10):
            result = self.run_client(USERNAME, PASSWORD, "send", "0.5", "0.5")
            self.assertEqual(result.returncode, 0, result.stderr)
            if cursor.is_visible():
                break
            time.sleep(0.2)
        expect(cursor).to_be_visible()

    def test_logout_ends_the_session(self) -> None:
        self.server.signup(USERNAME)
        self.login("/tables")
        expect(self.page.get_by_text("No items found.")).to_be_visible()
        self.assertEqual(self.server.sessions(), 1)

        self.page.get_by_role("button", name=f"Open account actions for {USERNAME}").click()
        self.page.get_by_role("button", name="Log out").click()

        expect(self.page).to_have_url(re.compile(r"/auth$"))
        # The app signs out at once and ends the session on the server in the background.
        self.wait_for(lambda: self.server.sessions() == 0, "the session was not ended")


if __name__ == "__main__":
    harness.main()
