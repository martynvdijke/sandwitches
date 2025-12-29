from sandwitches.models import Recipe, Rating
import pytest
from django.contrib.auth.models import User
from django.urls import reverse


@pytest.mark.django_db
def test_index_view_with_recipes(client, db):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    Recipe.objects.create(
        title="Test Recipe",
        description="This is a test recipe.",
        ingredients="Ingredient 1, Ingredient 2",
        instructions="Step 1, Step 2",
    )
    response = client.get("/recipes/test-recipe/", follow=True)
    assert response.status_code == 200
    assert b"Test Recipe" in response.content
    assert b"This is a test recipe." in response.content
    assert b"Ingredient 1" in response.content
    assert b"Ingredient 2" in response.content
    assert b"Step 1" in response.content
    assert b"Step 2" in response.content


@pytest.mark.django_db
def test_anonymous_cannot_rate_recipe(client):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    recipe = Recipe.objects.create(
        title="Rate Me",
        description="Rate this",
        ingredients="",
        instructions="",
    )
    resp = client.post(reverse("recipe_rate", kwargs={"pk": recipe.pk}))
    assert resp.status_code == 302
    # login_required should redirect to login page
    assert "/accounts/login/" in resp.url or "/login/" in resp.url


@pytest.mark.django_db
def test_logged_in_user_can_create_and_update_rating(client):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    user = User.objects.create_user("rater", "rater@example.com", "pw12345")
    recipe = Recipe.objects.create(
        title="Rateable",
        description="Rateable",
        ingredients="",
        instructions="",
    )

    # create rating
    assert Rating.objects.filter(recipe=recipe, user=user).count() == 0
    assert client.login(username="rater", password="pw12345")
    resp = client.post(
        reverse("recipe_rate", kwargs={"pk": recipe.pk}), {"score": "5"}, follow=True
    )
    assert resp.status_code == 200
    rating = Rating.objects.get(recipe=recipe, user=user)
    assert rating.score == 5
    assert b"Average" in resp.content and b"5.0" in resp.content
    assert b"Your rating: 5" in resp.content

    # update rating
    resp2 = client.post(
        reverse("recipe_rate", kwargs={"pk": recipe.pk}), {"score": "3"}, follow=True
    )
    assert resp2.status_code == 200
    rating.refresh_from_db()
    assert rating.score == 3
    detail = client.get(reverse("recipe_detail", kwargs={"slug": recipe.slug}))
    assert b"Your rating: 3" in detail.content


@pytest.mark.django_db
def test_multiple_users_affect_average_and_count(client):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    u1 = User.objects.create_user("u1", "u1@example.com", "pw1")  # noqa: F841
    u2 = User.objects.create_user("u2", "u2@example.com", "pw2")  # noqa: F841
    recipe = Recipe.objects.create(
        title="ManyRaters",
        description="ManyRaters",
        ingredients="",
        instructions="",
    )

    client.login(username="u1", password="pw1")
    client.post(reverse("recipe_rate", kwargs={"pk": recipe.pk}), {"score": "5"})
    client.logout()

    client.login(username="u2", password="pw2")
    client.post(reverse("recipe_rate", kwargs={"pk": recipe.pk}), {"score": "3"})
    client.logout()

    detail = client.get(reverse("recipe_detail", kwargs={"slug": recipe.slug}))
    # average of 5 and 3 is 4.0
    assert b"4.0" in detail.content
    # two ratings present
    assert b"2" in detail.content


@pytest.mark.django_db
def test_invalid_rating_rejected(client):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    user = User.objects.create_user("rater2", "r2@example.com", "pw222")
    recipe = Recipe.objects.create(
        title="InvalidRate",
        description="Invalid",
        ingredients="",
        instructions="",
    )
    client.login(username="rater2", password="pw222")
    resp = client.post(  # noqa: F841
        reverse("recipe_rate", kwargs={"pk": recipe.pk}),
        {"score": "11"},
        follow=True,
    )
    assert Rating.objects.filter(recipe=recipe, user=user).count() == 0
