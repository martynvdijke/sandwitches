import pytest
from django.core.exceptions import ValidationError
from sandwitches.models import Order, OrderItem, Recipe, User


@pytest.mark.django_db
def test_daily_order_limit():
    user = User.objects.create_user("limituser", "test@example.com", "password")
    recipe = Recipe.objects.create(
        title="Limited Recipe", price=10.00, uploaded_by=user, max_daily_orders=2
    )

    # First order
    o1 = Order.objects.create(user=user)
    OrderItem.objects.create(order=o1, recipe=recipe)
    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 1

    # Second order
    o2 = Order.objects.create(user=user)
    OrderItem.objects.create(order=o2, recipe=recipe)
    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 2

    # Third order should fail
    o3 = Order.objects.create(user=user)
    with pytest.raises(ValidationError, match="Daily order limit reached"):
        OrderItem.objects.create(order=o3, recipe=recipe)

    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 2
