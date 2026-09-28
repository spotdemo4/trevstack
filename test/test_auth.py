"""The client cannot add numbers without valid credentials."""

import harness

USERNAME = "demo_user"
PASSWORD = "example-password"


class AuthTest(harness.IntegrationTestCase):
    def test_rejects_unknown_user(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "visits", 7)

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.numbers(), [])

    def test_rejects_wrong_password(self) -> None:
        self.server.signup(USERNAME, PASSWORD)

        result = self.run_client(USERNAME, "wrong-password", "visits", 7)

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.numbers(), [])


if __name__ == "__main__":
    harness.main()
