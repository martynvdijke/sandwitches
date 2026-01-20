import pytest
from sandwitches.models import Order, Recipe, User


@pytest.mark.django_db
def test_order_creation(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(title="Priced Recipe", price=10.50, uploaded_by=user)

    order = Order.objects.create(user=user, recipe=recipe)

    assert order.total_price == 10.50
    assert order.status == "PENDING"
    assert str(order) == f"Order #{order.pk} - {user} - {recipe}"


@pytest.mark.django_db
def test_order_no_price_error(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(title="Free Recipe", uploaded_by=user)  # No price

    with pytest.raises(ValueError, match="Cannot order a recipe without a price"):
        Order.objects.create(user=user, recipe=recipe)


@pytest.mark.django_db
def test_order_price_snapshot(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(
        title="Inflation Recipe", price=5.00, uploaded_by=user
    )

    order = Order.objects.create(user=user, recipe=recipe)
    assert order.total_price == 5.00

    # Change recipe price
    recipe.price = 20.00
    recipe.save()

    # Order price should remain the same
    order.refresh_from_db()
    assert order.total_price == 5.00
