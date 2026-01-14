import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe

User = get_user_model()


@pytest.mark.django_db
def test_favorites_view_requires_login(client):
    url = reverse("favorites")
    response = client.get(url)
    assert response.status_code == 302
    assert "login" in response.url


@pytest.mark.django_db
def test_favorites_view_shows_favorited_recipes(client):
    user = User.objects.create_user(username="testuser", password="password")
    client.force_login(user)

    recipe1 = Recipe.objects.create(title="Recipe 1", description="Desc 1")
    Recipe.objects.create(title="Recipe 2", description="Desc 2")

    user.favorites.add(recipe1)

    url = reverse("favorites")
    response = client.get(url)

    assert response.status_code == 200
    assert "Recipe 1" in response.content.decode()
    assert "Recipe 2" not in response.content.decode()


@pytest.mark.django_db
def test_favorites_view_filtering(client):
    user = User.objects.create_user(username="testuser", password="password")
    client.force_login(user)

    recipe1 = Recipe.objects.create(title="Apple Pie", description="Sweet")
    recipe2 = Recipe.objects.create(title="Banana Bread", description="Sweet")

    user.favorites.add(recipe1)
    user.favorites.add(recipe2)

    url = reverse("favorites")
    response = client.get(url, {"q": "Apple"})

    assert response.status_code == 200
    assert "Apple Pie" in response.content.decode()
    assert "Banana Bread" not in response.content.decode()
