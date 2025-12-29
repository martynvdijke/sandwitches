import pytest
from django.urls import reverse
from sandwitches.models import Recipe, Rating

@pytest.mark.django_db
def test_ping_api(client):
    url = "/api/ping"
    response = client.get(url)
    assert response.status_code == 200
    assert response.json() == {"status": "ok", "message": "pong"}

@pytest.mark.django_db
def test_recipe_rating_api(client, user_factory):
    recipe = Recipe.objects.create(title="API Rate Test")
    u1 = user_factory(username="api1")
    u2 = user_factory(username="api2")
    
    Rating.objects.create(recipe=recipe, user=u1, score=9.0)
    Rating.objects.create(recipe=recipe, user=u2, score=7.0)
    
    url = f"/api/v1/recipes/{recipe.id}/rating"
    response = client.get(url)
    
    assert response.status_code == 200
    data = response.json()
    assert data["average"] == 8.0
    assert data["count"] == 2

@pytest.mark.django_db
def test_recipe_rating_api_404(client):
    url = "/api/v1/recipes/9999/rating"
    response = client.get(url)
    assert response.status_code == 404
