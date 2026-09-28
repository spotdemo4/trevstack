"""The client adds numbers through the server and they are stored in the database."""

import harness

USERNAME = "demo_user"
PASSWORD = "example-password"


class AddTest(harness.IntegrationTestCase):
    def setUp(self) -> None:
        super().setUp()
        self.server.signup(USERNAME, PASSWORD)

    def test_adds_number(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "visits", 7)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(result.stdout, "7\n")
        self.assertEqual(self.server.numbers(), [("visits", 7)])

    def test_returns_running_sum(self) -> None:
        first = self.run_client(USERNAME, PASSWORD, "visits", 7)
        second = self.run_client(USERNAME, PASSWORD, "signups", 5)

        self.assertEqual(first.returncode, 0, first.stderr)
        self.assertEqual(second.returncode, 0, second.stderr)
        self.assertEqual(first.stdout, "7\n")
        self.assertEqual(second.stdout, "12\n")
        self.assertEqual(self.server.numbers(), [("visits", 7), ("signups", 5)])

    def test_username_is_case_insensitive(self) -> None:
        result = self.run_client(USERNAME.upper(), PASSWORD, "visits", 7)

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.server.numbers(), [("visits", 7)])

    def test_rejects_number_out_of_range(self) -> None:
        result = self.run_client(USERNAME, PASSWORD, "visits", 0)

        self.assertNotEqual(result.returncode, 0)
        self.assertEqual(result.stdout, "")
        self.assertEqual(self.server.numbers(), [])


if __name__ == "__main__":
    harness.main()
