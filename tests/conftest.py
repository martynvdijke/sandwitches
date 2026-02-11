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
            import uuid

            kwargs["username"] = f"user_{uuid.uuid4().hex[:8]}"
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


@pytest.fixture
def tag_factory(db):
    def create_tag(**kwargs):
        from sandwitches.models import Tag

        if "name" not in kwargs:
            import uuid

            kwargs["name"] = f"Tag-{uuid.uuid4().hex[:6]}"
        return Tag.objects.create(**kwargs)

    return create_tag


@pytest.fixture
def recipe_factory(db, user_factory):
    def create_recipe(**kwargs):
        from sandwitches.models import Recipe

        if "uploaded_by" not in kwargs:
            kwargs["uploaded_by"] = user_factory()
        if "title" not in kwargs:
            import uuid

            kwargs["title"] = f"Recipe-{uuid.uuid4().hex[:6]}"
        if "price" not in kwargs:
            kwargs["price"] = 10.0
        return Recipe.objects.create(**kwargs)

    return create_recipe


@pytest.fixture
def rating_factory(db, user_factory, recipe_factory):
    def create_rating(**kwargs):
        from sandwitches.models import Rating

        if "user" not in kwargs:
            kwargs["user"] = user_factory()
        if "recipe" not in kwargs:
            kwargs["recipe"] = recipe_factory()
        if "score" not in kwargs:
            kwargs["score"] = 5.0
        return Rating.objects.create(**kwargs)

    return create_rating


@pytest.fixture
def order_factory(db, user_factory):
    def create_order(**kwargs):
        from sandwitches.models import Order

        if "user" not in kwargs:
            kwargs["user"] = user_factory()
        return Order.objects.create(**kwargs)

    return create_order


@pytest.fixture(autouse=True)
def override_static_storage(settings):
    """
    Use standard StaticFilesStorage for tests to avoid manifest missing errors
    with WhiteNoise's CompressedManifestStaticFilesStorage.
    """
    settings.STORAGES = {
        "default": {
            "BACKEND": "django.core.files.storage.FileSystemStorage",
        },
        "staticfiles": {
            "BACKEND": "django.contrib.staticfiles.storage.StaticFilesStorage",
        },
    }
