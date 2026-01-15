# conftest.py
import os

os.environ["DJANGO_ALLOW_ASYNC_UNSAFE"] = "true"

import pytest


# @pytest.fixture()
# def setup_environment():
#     """Set up environment variables for testing."""
#     if not settings.configured:
#         # os.environ["DJANGO_SETTINGS_MODULE"] = "sandwitches.sandwitches.settings"
#         os.environ["DEBUG"] = "TRUE"
#         os.environ["SECRET_KEY"] = "test_secret_key"
#         os.environ["ALLOWED_HOSTS"] = "127.0.0.1"
#         os.environ["CSRF_TRUSTED_ORIGINS"] = "http://127.0.0.1"
#         settings.configure()
#         setup_test_environment()


@pytest.fixture
def client(client):
    """
    A standard Django test client fixture.
    The 'client' fixture is provided by pytest-django.
    """
    return client


@pytest.fixture
def user_factory(db):
    """
    A factory function to create User objects for tests.
    The 'db' fixture ensures the database is ready.
    """

    def create_user(**kwargs):
        from django.contrib.auth import get_user_model

        User = get_user_model()
        # Set reasonable defaults
        if "username" not in kwargs:
            kwargs["username"] = "testuser"
        if "password" not in kwargs:
            kwargs["password"] = "testpass123"

        return User.objects.create_user(**kwargs)

    return create_user


@pytest.fixture
def staff_user(user_factory):
    """
    Creates and returns a staff user.
    """
    return user_factory(username="staff_test", is_staff=True, email="staff@example.com")
