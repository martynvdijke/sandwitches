import pytest
from sandwitches.models import Recipe, Tag, Rating, Order, CartItem


@pytest.mark.django_db
class TestAPIPermissions:
    def test_recipes_list_public(self, client, recipe_factory):
        recipe_factory(title="Recipe 1")
        response = client.get("/api/v1/recipes")
        assert response.status_code == 200
        assert len(response.json()) >= 1

    def test_create_recipe_requires_auth(self, client):
        data = {"title": "New Recipe", "description": "Desc"}
        response = client.post(
            "/api/v1/recipes", data=data, content_type="application/json"
        )
        assert response.status_code == 401

    def test_create_recipe_authenticated(self, admin_client):
        data = {
            "title": "New Recipe",
            "description": "Desc",
            "servings": 2,
            "price": 10.50,
        }
        response = admin_client.post(
            "/api/v1/recipes", data=data, content_type="application/json"
        )
        assert response.status_code == 201
        assert Recipe.objects.filter(title="New Recipe").exists()

    def test_tags_list_public(self, client):
        Tag.objects.create(name="Spicy")
        response = client.get("/api/v1/tags")
        assert response.status_code == 200
        assert any(tag["name"] == "Spicy" for tag in response.json())

    def test_create_tag_requires_staff(self, client, user_factory):
        user = user_factory()
        client.force_login(user)
        data = {"name": "Staff Only"}
        response = client.post(
            "/api/v1/tags", data=data, content_type="application/json"
        )
        assert response.status_code == 403

    def test_create_tag_staff(self, admin_client):
        data = {"name": "New Tag"}
        response = admin_client.post(
            "/api/v1/tags", data=data, content_type="application/json"
        )
        assert response.status_code == 201
        assert Tag.objects.filter(name="New Tag").exists()

    def test_ratings_create_requires_auth(self, client, recipe_factory):
        recipe = recipe_factory()
        data = {"score": 5, "comment": "Good"}
        response = client.post(
            f"/api/v1/recipes/{recipe.id}/ratings",
            data=data,
            content_type="application/json",
        )
        assert response.status_code == 401

    def test_ratings_create_authenticated(self, client, user_factory, recipe_factory):
        user = user_factory()
        recipe = recipe_factory()
        client.force_login(user)
        data = {"score": 8, "comment": "Yum"}
        response = client.post(
            f"/api/v1/recipes/{recipe.id}/ratings",
            data=data,
            content_type="application/json",
        )
        assert response.status_code == 201
        assert Rating.objects.filter(recipe=recipe, user=user, score=8).exists()

    def test_cart_requires_auth(self, client):
        response = client.get("/api/v1/cart")
        assert response.status_code == 401

    def test_cart_operations(self, client, user_factory, recipe_factory):
        user = user_factory()
        recipe = recipe_factory()
        client.force_login(user)

        # Add to cart
        response = client.post(
            "/api/v1/cart",
            data={"recipe_id": recipe.id, "quantity": 2},
            content_type="application/json",
        )
        assert response.status_code == 201
        assert CartItem.objects.filter(user=user, recipe=recipe, quantity=2).exists()

        item_id = response.json()["id"]

        # Update cart
        response = client.patch(
            f"/api/v1/cart/{item_id}",
            data={"quantity": 5},
            content_type="application/json",
        )
        assert response.status_code == 200
        assert CartItem.objects.get(id=item_id).quantity == 5

        # Delete cart item
        response = client.delete(f"/api/v1/cart/{item_id}")
        assert response.status_code == 204
        assert not CartItem.objects.filter(id=item_id).exists()

    def test_orders_list_permissions(
        self, client, user_factory, admin_client, recipe_factory
    ):
        user = user_factory()
        recipe_factory()
        Order.objects.create(user=user)

        # Anonymous
        response = client.get("/api/v1/orders")
        assert response.status_code == 401

        # Owner
        client.force_login(user)
        response = client.get("/api/v1/orders")
        assert response.status_code == 200
        assert len(response.json()) == 1

        # Other user
        other_user = user_factory(username="other")
        client.force_login(other_user)
        response = client.get("/api/v1/orders")
        assert response.status_code == 200
        assert len(response.json()) == 0

        # Staff
        response = admin_client.get("/api/v1/orders")
        assert response.status_code == 200
        assert len(response.json()) >= 1
