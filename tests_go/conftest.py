import os
import signal
import subprocess
import tempfile
import time
import urllib.request

import pytest

GO_BINARY = os.path.join(os.path.dirname(__file__), "..", "go-app", "sandwitches-go")


class GoServer:
    """Manages a Go sandwitches server process for testing."""

    def __init__(self):
        self.process = None
        self.db_file = None
        self.media_dir = None

    def start(self):
        self.db_file = os.path.join(tempfile.mkdtemp(), "sandwitches.db")
        self.media_dir = tempfile.mkdtemp()
        env = {
            **os.environ,
            "DATABASE_FILE": self.db_file,
            "MEDIA_ROOT": self.media_dir,
            "SECRET_KEY": "test-secret-key-12345",
            "DEBUG": "true",
            "Django_DB_PATH": "/nonexistent/__skip_django_migration__.db",
            "PORT": "6279",
        }
        self.process = subprocess.Popen(
            [GO_BINARY],
            env=env,
            cwd=os.path.join(os.path.dirname(__file__), "..", "go-app"),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
        )
        self._wait_until_ready()

    def _wait_until_ready(self, timeout=15):
        deadline = time.time() + timeout
        while time.time() < deadline:
            if self.process.poll() is not None:
                out, err = self.process.communicate()
                raise RuntimeError(
                    f"Go server exited early: {self.process.returncode}\n"
                    f"stdout: {out.decode()}\nstderr: {err.decode()}"
                )
            try:
                urllib.request.urlopen("http://127.0.0.1:6279/", timeout=2)
                return
            except Exception:
                time.sleep(0.3)
        raise RuntimeError("Go server did not start within timeout")

    def stop(self):
        if self.process:
            self.process.send_signal(signal.SIGTERM)
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()

    @property
    def url(self):
        return "http://127.0.0.1:6279"


def _reset_db(server):
    """Tear down and restart the server with a fresh database."""
    server.stop()
    if os.path.exists(server.db_file):
        os.remove(server.db_file)
    if server.media_dir and os.path.exists(server.media_dir):
        import shutil
        shutil.rmtree(server.media_dir, ignore_errors=True)
    server.start()


@pytest.fixture(scope="session")
def go_server():
    """Session-scoped Go server fixture."""
    if not os.path.exists(GO_BINARY):
        pytest.skip(f"Go binary not found at {GO_BINARY}")
    server = GoServer()
    server.start()
    yield server
    server.stop()


@pytest.fixture(autouse=True)
def fresh_db(go_server):
    """Reset the database before each test for isolation."""
    _reset_db(go_server)
    yield


def create_admin(go_server, username="admin", password="adminpass123"):
    """Create the initial admin via the setup page."""
    import http.client
    from urllib.parse import urlencode
    parsed = urllib.request.urlparse(go_server.url)
    conn = http.client.HTTPConnection(parsed.hostname, parsed.port)
    data = urlencode({
        "username": username,
        "email": f"{username}@test.com",
        "password1": password,
        "password2": password,
    })
    conn.request("POST", "/setup", body=data, headers={"Content-Type": "application/x-www-form-urlencoded"})
    resp = conn.getresponse()
    resp.read()
    # Setup redirects to /dashboard/ after success (302)
    return resp.status == 302


def login_session(page, go_server, username, password):
    """Log in through the UI and return the page after redirect."""
    page.goto(f"{go_server.url}/login")
    page.fill("input[name='username']", username)
    page.fill("input[name='password']", password)
    page.press("input[name='password']", "Enter")
    # Wait for redirect to / 
    page.wait_for_url(go_server.url + "/", timeout=5000)


def create_recipe_via_ui(page, go_server, title, description, ingredients, instructions, servings=2, price=None):
    """Create a recipe via the admin dashboard (requires logged-in admin)."""
    page.goto(f"{go_server.url}/dashboard/recipes/add")
    # Wait for the form title heading
    page.wait_for_selector("h4", timeout=10000)
    page.fill("input[name='title']", title)
    page.fill("textarea[name='description']", description)
    page.fill("textarea[name='ingredients']", ingredients)
    page.fill("textarea[name='instructions']", instructions)
    page.fill("input[name='servings']", str(servings))
    if price is not None:
        page.fill("input[name='price']", price)
    # Enable is_approved checkbox for admin-added recipes
    checkbox = page.locator("input[name='is_approved']")
    if checkbox.count():
        checkbox.check()
    page.locator("button[type='submit']").click()
    # Should redirect to recipe list
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=5000)
