"""The client cannot send positions without valid credentials."""

import harness

USERNAME = "demo_user"
PASSWORD = harness.PASSWORD


class AuthTest(harness.IntegrationTestCase):
    def test_rejects_unknown_user(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "send", "0.5", "0.5")

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.cursors(), [])

    def test_rejects_wrong_password(self) -> None:
        self.server.signup(USERNAME)

        result = self.run_client(USERNAME, "wrong-password", "send", "0.5", "0.5")

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.cursors(), [])


if __name__ == "__main__":
    harness.main()
