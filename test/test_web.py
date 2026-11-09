"""The web app signs users up and in, lists the numbers they add, and signs them out."""

import re
import time

import harness
from playwright.sync_api import expect

USERNAME = "demo_user"
PASSWORD = harness.PASSWORD


class WebTest(harness.BrowserTestCase):
    def test_redirects_to_sign_in(self) -> None:
        self.page.goto("/numbers")

        expect(self.page).to_have_url(re.compile(r"/auth\?"))
        expect(self.page.get_by_role("heading", name="Welcome to TrevStack")).to_be_visible()

    def test_signup_then_list_numbers(self) -> None:
        self.page.goto("/numbers")
        self.page.get_by_role("tab", name="Sign up").click()
        signup = self.page.get_by_role("tabpanel", name="Sign up")
        signup.get_by_label("Username").fill(USERNAME)
        signup.get_by_label("Password").fill(PASSWORD)
        signup.get_by_role("button", name="Create account").click()

        expect(self.page).to_have_url(re.compile(r"/numbers$"))
        expect(self.page.get_by_text("No items found.")).to_be_visible()

        result = self.run_client(USERNAME, PASSWORD, "visits", 7)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.page.reload()

        row = self.page.get_by_role("row").filter(has_text="visits")
        expect(row).to_contain_text("7")

    def test_logout_ends_the_session(self) -> None:
        self.server.signup(USERNAME)
        self.page.goto("/numbers")
        login = self.page.get_by_role("tabpanel", name="Login")
        login.get_by_label("Username").fill(USERNAME)
        login.get_by_label("Password").fill(PASSWORD)
        login.get_by_role("button", name="Login").click()
        expect(self.page).to_have_url(re.compile(r"/numbers$"))
        expect(self.page.get_by_text("No items found.")).to_be_visible()
        self.assertEqual(self.server.sessions(), 1)

        self.page.get_by_role("button", name=f"Open account actions for {USERNAME}").click()
        self.page.get_by_role("button", name="Log out").click()

        expect(self.page).to_have_url(re.compile(r"/auth$"))
        # The app signs out at once and ends the session on the server in the background.
        deadline = time.monotonic() + harness.BROWSER_TIMEOUT / 1000
        while self.server.sessions() and time.monotonic() < deadline:
            time.sleep(0.05)
        self.assertEqual(self.server.sessions(), 0)


if __name__ == "__main__":
    harness.main()
