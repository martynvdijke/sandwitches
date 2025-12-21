from sandwitches.models import Recipe
import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model

User = get_user_model()


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


@pytest.mark.django_db
def test_setup_page_shown_when_no_superuser(client):
    resp = client.get(reverse("setup"))
    assert resp.status_code == 200


@pytest.mark.django_db
def test_setup_redirects_when_superuser_exists(client):
    User.objects.create_superuser("admin", "admin@example.com", "pw")
    resp = client.get(reverse("setup"))
    assert resp.status_code in (302, 301)
    assert resp.url == reverse("index") or resp.url.endswith("/")


@pytest.mark.django_db
def test_setup_creates_superuser_and_logs_in(client):
    data = {
        "username": "siteadmin",
        "email": "siteadmin@example.com",
        "first_name": "Site",
        "last_name": "Admin",
        "password1": "strong-pass-123",
        "password2": "strong-pass-123",
    }
    resp = client.post(reverse("setup"), data, follow=True)
    assert resp.status_code == 200
    u = User.objects.get(username="siteadmin")
    assert u.is_superuser and u.is_staff
    # the user should be logged in after setup
    assert resp.context["user"].is_authenticated
