import pytest
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe, Rating

User = get_user_model()


@pytest.mark.django_db
def test_recipe_rating_helpers():
    r = Recipe.objects.create(
        title="R1", description="", ingredients="", instructions=""
    )
    u1 = User.objects.create_user("u1", "u1@example.com", "p1")
    u2 = User.objects.create_user("u2", "u2@example.com", "p2")
    Rating.objects.create(recipe=r, user=u1, score=5)
    Rating.objects.create(recipe=r, user=u2, score=3)
    # average should be 4.0 and count 2
    avg = r.average_rating()
    assert round(float(avg), 1) == 4.0
    assert r.rating_count() == 2
