from sandwitches.models import Recipe
import pytest
from django.contrib.auth.models import User


@pytest.mark.django_db
def test_index_view_redirects_to_setup(client):
    response = client.get("/")
    assert response.status_code == 302
    assert response.url.endswith("/setup/")


@pytest.mark.django_db
def test_index(client, db):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    response = client.get("/")
    assert response.status_code == 200
    assert b"No sandwitches yet, please stay tuned." in response.content


@pytest.mark.django_db
def test_index_view_with_recipes(client, db):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    Recipe.objects.create(
        title="Test Recipe",
        description="This is a test recipe.",
        ingredients="Ingredient 1, Ingredient 2",
        instructions="Step 1, Step 2",
    )
    response = client.get("/")
    assert response.status_code == 200
    assert b"Test Recipe" in response.content
