import os
import signal
import subprocess
import tempfile
import time
import urllib.request

import pytest

GO_BINARY = os.path.join(os.path.dirname(__file__), "..", "go-app", "sandwitches-go")


class GoServer:
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
    server.stop()
    if os.path.exists(server.db_file):
        os.remove(server.db_file)
    if server.media_dir and os.path.exists(server.media_dir):
        import shutil
        shutil.rmtree(server.media_dir, ignore_errors=True)
    server.start()


@pytest.fixture(scope="session")
def go_server():
    if not os.path.exists(GO_BINARY):
        pytest.skip(f"Go binary not found at {GO_BINARY}")
    server = GoServer()
    server.start()
    yield server
    server.stop()


@pytest.fixture(autouse=True)
def fresh_db(go_server):
    _reset_db(go_server)
    yield


def create_admin(go_server, username="admin", password="adminpass123"):
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
    return resp.status == 302


def login_session(page, go_server, username, password):
    page.goto(f"{go_server.url}/login", wait_until="commit", timeout=15000)
    page.wait_for_selector("input[name='password']", timeout=10000)
    page.fill("input[name='username']", username)
    page.fill("input[name='password']", password)
    page.locator("button[type='submit']").click(no_wait_after=True, timeout=5000)
    page.wait_for_url(go_server.url + "/", wait_until="commit", timeout=10000)


def create_recipe_via_ui(page, go_server, title, description, ingredients, instructions, servings=2, price=None):
    page.goto(f"{go_server.url}/dashboard/recipes/add")
    page.wait_for_selector("h4", timeout=10000)
    page.fill("input[name='title']", title)
    page.fill("textarea[name='description']", description)
    page.fill("textarea[name='ingredients']", ingredients)
    page.fill("textarea[name='instructions']", instructions)
    page.fill("input[name='servings']", str(servings))
    if price is not None:
        page.fill("input[name='price']", price)
    checkbox = page.locator("input[name='is_approved']")
    if checkbox.count():
        checkbox.check()
    page.locator("button[type='submit']").click()
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=5000)


def get_slug_from_admin_list(page, go_server):
    page.goto(f"{go_server.url}/dashboard/recipes")
    href = page.locator("tbody td a").first.get_attribute("href") or ""
    return href.rstrip("/").rsplit("/", 1)[-1]
