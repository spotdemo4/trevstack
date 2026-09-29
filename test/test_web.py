"""The web app signs users up and lists the numbers they add."""

import re

import harness
from playwright.sync_api import expect

USERNAME = "demo_user"
PASSWORD = "example-password"


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


if __name__ == "__main__":
    harness.main()
