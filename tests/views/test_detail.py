from sandwitches.models import Recipe
import pytest
from django.contrib.auth.models import User


@pytest.mark.django_db
def test_index_view_with_recipes(client, db):
    User.objects.create_superuser("admin", "admin@example.com", "strongpassword123")
    Recipe.objects.create(
        title="Test Recipe",
        description="This is a test recipe.",
        ingredients="Ingredient 1, Ingredient 2",
        instructions="Step 1, Step 2",
    )
    response = client.get("/recipes/test-recipe/")
    assert response.status_code == 200
    assert b"Test Recipe" in response.content
    assert b"This is a test recipe." in response.content
    assert b"Ingredient 1" in response.content
    assert b"Ingredient 2" in response.content
    assert b"Step 1" in response.content
    assert b"Step 2" in response.content
