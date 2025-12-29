import pytest
from django.urls import reverse
from sandwitches.models import Recipe, Rating
from django.contrib.auth import get_user_model

User = get_user_model()


@pytest.fixture
def recipe():
    return Recipe.objects.create(title="Rate Me")


@pytest.mark.django_db
def test_rate_recipe_authenticated(client, user_factory, recipe):
    user = user_factory()
    client.force_login(user)

    url = reverse("recipe_rate", kwargs={"pk": recipe.pk})
    data = {"score": "5"}

    response = client.post(url, data, follow=True)

    assert response.status_code == 200
    assert Rating.objects.filter(recipe=recipe, user=user, score=5).exists()
    assert recipe.average_rating() == 5.0


@pytest.mark.django_db
def test_rate_recipe_update_existing(client, user_factory, recipe):
    user = user_factory()
    client.force_login(user)
    Rating.objects.create(recipe=recipe, user=user, score=3)

    url = reverse("recipe_rate", kwargs={"pk": recipe.pk})
    data = {"score": "1"}

    response = client.post(url, data, follow=True)

    assert response.status_code == 200
    r_obj = Rating.objects.get(recipe=recipe, user=user)
    assert r_obj.score == 1
    assert recipe.average_rating() == 1.0


@pytest.mark.django_db
def test_rate_recipe_unauthenticated(client, recipe):
    url = reverse("recipe_rate", kwargs={"pk": recipe.pk})
    data = {"score": "5"}

    response = client.post(url, data)

    # Should redirect to login
    assert response.status_code == 302
    assert "/accounts/login/" in response.url
    assert Rating.objects.count() == 0


@pytest.mark.django_db
def test_rate_recipe_invalid_method(client, user_factory, recipe):
    user = user_factory()
    client.force_login(user)
    url = reverse("recipe_rate", kwargs={"pk": recipe.pk})

    response = client.get(url)

    # Should redirect back to detail
    assert response.status_code == 302
    assert response.url == reverse("recipe_detail", kwargs={"slug": recipe.slug})


@pytest.mark.django_db
def test_rate_recipe_invalid_form(client, user_factory, recipe):
    user = user_factory()
    client.force_login(user)
    url = reverse("recipe_rate", kwargs={"pk": recipe.pk})

    # Missing score
    response = client.post(url, {}, follow=True)

    # Should show error message (checking messages framework is a bit verbose in raw client,
    # but we can check if rating was created)
    assert Rating.objects.count() == 0
    # Should redirect back to detail
    # We followed redirect, so status is 200, check template used maybe?
    assert "detail.html" in [t.name for t in response.templates]
