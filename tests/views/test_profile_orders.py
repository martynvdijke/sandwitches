import pytest
from django.urls import reverse
from sandwitches.models import Recipe, Order
from decimal import Decimal


@pytest.mark.django_db
def test_profile_shows_orders(client, user_factory):
    user = user_factory()
    client.force_login(user)

    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("5.00"), servings=1
    )
    order = Order.objects.create(
        user=user, recipe=recipe, total_price=Decimal("5.00"), status="PENDING"
    )

    url = reverse("user_profile")
    response = client.get(url)

    assert response.status_code == 200
    assert "orders" in response.context
    assert order in response.context["orders"]
    assert b"Test Sandwich" in response.content
    # Check for display value of status. PENDING display is "Pending"
    assert b"Pending" in response.content
    assert b"5.00" in response.content


@pytest.mark.django_db
def test_profile_order_filtering(client, user_factory):
    user = user_factory()
    client.force_login(user)

    r1 = Recipe.objects.create(title="S1", price=Decimal("1.00"))
    r2 = Recipe.objects.create(title="S2", price=Decimal("2.00"))

    Order.objects.create(
        user=user, recipe=r1, status="PENDING", total_price=Decimal("1.00")
    )
    Order.objects.create(
        user=user, recipe=r2, status="COMPLETED", total_price=Decimal("2.00")
    )

    url = reverse("user_profile")

    # Filter by COMPLETED
    response = client.get(f"{url}?status=COMPLETED")
    assert response.status_code == 200
    assert len(response.context["orders"]) == 1
    assert response.context["orders"][0].recipe.title == "S2"

    # Filter by PENDING
    response = client.get(f"{url}?status=PENDING")
    assert len(response.context["orders"]) == 1
    assert response.context["orders"][0].recipe.title == "S1"


@pytest.mark.django_db
def test_profile_order_sorting(client, user_factory):
    user = user_factory()
    client.force_login(user)

    r1 = Recipe.objects.create(title="A", price=Decimal("10.00"))
    r2 = Recipe.objects.create(title="B", price=Decimal("5.00"))

    Order.objects.create(user=user, recipe=r1, total_price=Decimal("10.00"))
    Order.objects.create(user=user, recipe=r2, total_price=Decimal("5.00"))

    url = reverse("user_profile")

    # Sort by price ascending
    response = client.get(f"{url}?sort=price_asc")
    orders = list(response.context["orders"])
    assert orders[0].total_price == Decimal("5.00")
    assert orders[1].total_price == Decimal("10.00")

    # Sort by price descending
    response = client.get(f"{url}?sort=price_desc")
    orders = list(response.context["orders"])
    assert orders[0].total_price == Decimal("10.00")
    assert orders[1].total_price == Decimal("5.00")
