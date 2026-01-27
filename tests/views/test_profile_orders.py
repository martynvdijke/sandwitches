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
