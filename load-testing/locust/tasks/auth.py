import random
from locust import TaskSet, task

TEST_USERS = [
    {"email": "user1@test.shopos.local", "password": "Password1!"},
    {"email": "user2@test.shopos.local", "password": "Password2!"},
    {"email": "user3@test.shopos.local", "password": "Password3!"},
    {"email": "user4@test.shopos.local", "password": "Password4!"},
    {"email": "user5@test.shopos.local", "password": "Password5!"},
]


class AuthTasks(TaskSet):
    @task(3)
    def login(self):
        user = random.choice(TEST_USERS)
        with self.client.post(
            "/api/v1/auth/login",
            json={"email": user["email"], "password": user["password"]},
            name="/api/v1/auth/login",
            catch_response=True,
        ) as resp:
            if resp.status_code == 200:
                token = resp.json().get("access_token")
                if token:
                    self.user.token = token
                    resp.success()
                else:
                    resp.failure("No access_token in response")
            elif resp.status_code == 401:
                resp.failure(f"Login 401 for {user['email']}")

    @task(1)
    def refresh_token(self):
        if not getattr(self.user, "token", None):
            return
        with self.client.post(
            "/api/v1/auth/refresh",
            json={"token": self.user.token},
            name="/api/v1/auth/refresh",
            catch_response=True,
        ) as resp:
            if resp.status_code == 200:
                self.user.token = resp.json().get("access_token", self.user.token)
                resp.success()
            else:
                resp.failure(f"Refresh failed: {resp.status_code}")

    @task(1)
    def logout(self):
        if not getattr(self.user, "token", None):
            return
        with self.client.post(
            "/api/v1/auth/logout",
            headers={"Authorization": f"Bearer {self.user.token}"},
            name="/api/v1/auth/logout",
            catch_response=True,
        ) as resp:
            if resp.status_code in (200, 204):
                self.user.token = None
                resp.success()
