import pytest
from django.urls import reverse
from sandwitches.models import Recipe, Order, OrderItem
from decimal import Decimal
import datetime
from django.utils import timezone


@pytest.mark.django_db
def test_profile_pagination(client, user_factory):
    user = user_factory()
    client.force_login(user)
    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("5.00"), servings=1
    )

    # Create 6 orders (page size is 5)
    for i in range(6):
        o = Order.objects.create(
            user=user, total_price=Decimal("5.00"), status="PENDING"
        )
        OrderItem.objects.create(order=o, recipe=recipe)

    url = reverse("user_profile")
    response = client.get(url)

    assert response.status_code == 200
    assert len(response.context["orders"]) == 5
    assert response.context["orders"].has_next()

    response = client.get(url + "?page=2")
    assert response.status_code == 200
    assert len(response.context["orders"]) == 1
    assert not response.context["orders"].has_next()


@pytest.mark.django_db
def test_profile_filtering(client, user_factory):
    user = user_factory()
    client.force_login(user)
    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("5.00"), servings=1
    )

    o1 = Order.objects.create(user=user, total_price=Decimal("5.00"), status="PENDING")
    OrderItem.objects.create(order=o1, recipe=recipe)

    o2 = Order.objects.create(
        user=user, total_price=Decimal("5.00"), status="COMPLETED"
    )
    OrderItem.objects.create(order=o2, recipe=recipe)

    url = reverse("user_profile")

    # Filter by PENDING
    response = client.get(url, {"status": "PENDING"})
    assert response.status_code == 200
    assert len(response.context["orders"]) == 1
    assert response.context["orders"][0].status == "PENDING"

    # Filter by COMPLETED
    response = client.get(url, {"status": "COMPLETED"})
    assert response.status_code == 200
    assert len(response.context["orders"]) == 1
    assert response.context["orders"][0].status == "COMPLETED"


@pytest.mark.django_db
def test_profile_sorting(client, user_factory):
    user = user_factory()
    client.force_login(user)
    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("5.00"), servings=1
    )

    order1 = Order.objects.create(
        user=user, total_price=Decimal("10.00"), status="PENDING"
    )
    OrderItem.objects.create(order=order1, recipe=recipe, price=Decimal("10.00"))
    order1.created_at = timezone.now() - datetime.timedelta(days=2)
    order1.save()

    order2 = Order.objects.create(
        user=user, total_price=Decimal("5.00"), status="PENDING"
    )
    OrderItem.objects.create(order=order2, recipe=recipe, price=Decimal("5.00"))
    order2.created_at = timezone.now() - datetime.timedelta(days=1)
    order2.save()

    url = reverse("user_profile")

    # Sort by date asc (oldest first)
    response = client.get(url, {"sort": "date_asc"})
    assert response.status_code == 200
    assert response.context["orders"][0] == order1
    assert response.context["orders"][1] == order2

    # Sort by date desc (newest first)
    response = client.get(url, {"sort": "date_desc"})
    assert response.status_code == 200
    assert response.context["orders"][0] == order2
    assert response.context["orders"][1] == order1

    # Sort by price asc
    response = client.get(url, {"sort": "price_asc"})
    assert response.status_code == 200
    assert response.context["orders"][0] == order2  # 5.00
    assert response.context["orders"][1] == order1  # 10.00

    # Sort by price desc
    response = client.get(url, {"sort": "price_desc"})
    assert response.status_code == 200
    assert response.context["orders"][0] == order1  # 10.00
    assert response.context["orders"][1] == order2  # 5.00


@pytest.mark.django_db
def test_user_order_detail(client, user_factory):
    user = user_factory()
    client.force_login(user)
    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("5.00"), servings=1
    )
    order = Order.objects.create(
        user=user, total_price=Decimal("5.00"), status="PENDING"
    )
    OrderItem.objects.create(order=order, recipe=recipe)

    url = reverse("user_order_detail", kwargs={"pk": order.pk})
    response = client.get(url)

    assert response.status_code == 200
    assert response.context["order"] == order
    assert b"Order Details" in response.content
