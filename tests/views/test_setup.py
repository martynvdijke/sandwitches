import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model

User = get_user_model()


@pytest.mark.django_db
def test_setup_access_no_superuser(client):
    # Ensure no superuser
    User.objects.filter(is_superuser=True).delete()

    url = reverse("setup")
    response = client.get(url)
    assert response.status_code == 200
    assert "setup.html" in [t.name for t in response.templates]


@pytest.mark.django_db
def test_setup_access_with_superuser(client, user_factory):
    # Create superuser
    user_factory(is_superuser=True)

    url = reverse("setup")
    response = client.get(url)

    # Should redirect to index
    assert response.status_code == 302
    assert response.url == reverse("index")


@pytest.mark.django_db
def test_setup_creates_superuser(client):
    User.objects.filter(is_superuser=True).delete()

    url = reverse("setup")
    data = {
        "username": "admin",
        "email": "admin@example.com",
        "password1": "password123",
        "password2": "password123",
    }

    response = client.post(url, data, follow=True)

    assert User.objects.filter(username="admin", is_superuser=True).exists()
    assert response.status_code == 200
    # Should redirect to admin index
    assert response.redirect_chain[-1][0] == reverse("admin:index")


@pytest.mark.django_db
def test_index_redirects_to_setup_if_no_superuser(client):
    User.objects.filter(is_superuser=True).delete()

    url = reverse("index")
    response = client.get(url)

    assert response.status_code == 302
    assert response.url == reverse("setup")


@pytest.mark.django_db
def test_index_renders_if_superuser_exists(client, user_factory):
    user_factory(is_superuser=True)

    url = reverse("index")
    response = client.get(url)

    assert response.status_code == 200
    assert "index.html" in [t.name for t in response.templates]
