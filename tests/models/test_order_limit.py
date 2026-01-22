import pytest
from django.core.exceptions import ValidationError
from sandwitches.models import Order, Recipe, User


@pytest.mark.django_db
def test_daily_order_limit():
    user = User.objects.create_user("limituser", "test@example.com", "password")
    recipe = Recipe.objects.create(
        title="Limited Recipe", price=10.00, uploaded_by=user, max_daily_orders=2
    )

    # First order
    Order.objects.create(user=user, recipe=recipe)
    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 1

    # Second order
    Order.objects.create(user=user, recipe=recipe)
    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 2

    # Third order should fail
    with pytest.raises(ValidationError, match="Daily order limit reached"):
        Order.objects.create(user=user, recipe=recipe)

    recipe.refresh_from_db()
    assert recipe.daily_orders_count == 2
