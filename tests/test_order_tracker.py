import pytest
from django.urls import reverse
from sandwitches.models import Order, Recipe, OrderItem
from django.contrib.auth import get_user_model

User = get_user_model()


@pytest.mark.django_db
def test_order_tracker_view(client):
    user = User.objects.create_user(username="testuser", password="password")
    recipe = Recipe.objects.create(title="Test Recipe", price=10.0)
    order = Order.objects.create(user=user, total_price=10.0)
    OrderItem.objects.create(order=order, recipe=recipe, quantity=1, price=10.0)

    url = reverse("order_tracker", kwargs={"token": order.tracking_token})
    response = client.get(url)

    assert response.status_code == 200
    assert str(order.id).encode() in response.content
    assert b"Track Order" in response.content
    assert b"Test Recipe" in response.content


@pytest.mark.django_db
def test_order_tracker_invalid_token(client):
    import uuid

    url = reverse("order_tracker", kwargs={"token": uuid.uuid4()})
    response = client.get(url)
    assert response.status_code == 404


@pytest.mark.django_db
def test_profile_page_contains_tracking_link(client):
    user = User.objects.create_user(username="testuser", password="password")
    client.login(username="testuser", password="password")
    recipe = Recipe.objects.create(title="Test Recipe", price=10.0)
    order = Order.objects.create(user=user, total_price=10.0)
    OrderItem.objects.create(order=order, recipe=recipe, quantity=1, price=10.0)

    url = reverse("user_profile")
    response = client.get(url)

    assert response.status_code == 200
    tracking_url = reverse("order_tracker", kwargs={"token": order.tracking_token})
    assert tracking_url.encode() in response.content
