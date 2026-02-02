import pytest
from sandwitches.models import Order, Recipe, User, OrderItem


@pytest.mark.django_db
def test_order_creation(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(title="Priced Recipe", price=10.50, uploaded_by=user)

    order = Order.objects.create(user=user)
    OrderItem.objects.create(order=order, recipe=recipe, quantity=1)
    order.total_price = recipe.price
    order.save()

    assert order.total_price == 10.50
    assert order.status == "PENDING"
    assert str(order) == f"Order #{order.pk} - {user}"
    assert order.items.count() == 1
    assert order.items.first().recipe == recipe


@pytest.mark.django_db
def test_order_no_price_error(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(title="Free Recipe", uploaded_by=user)  # No price

    # Validation is now in OrderItem or view.
    # If we create OrderItem directly, we can check logic there.
    # OrderItem.save() handles defaults, but if recipe has no price, it raises exception or fails?
    # Original logic: "Cannot order a recipe without a price" was in Order.save

    # Let's check OrderItem logic if it enforces price
    # The models.py I wrote:
    # if not self.price: self.price = self.recipe.price
    # If recipe.price is None, self.price becomes None.
    # price field is DecimalField, nullable?
    # models.py: price = models.DecimalField(max_digits=6, decimal_places=2)
    # It is NOT nullable (default). So it will raise IntegrityError or ValidationError on save if None.

    order = Order.objects.create(user=user)
    with pytest.raises(Exception):  # IntegrityError or similar
        OrderItem.objects.create(order=order, recipe=recipe)


@pytest.mark.django_db
def test_order_price_snapshot(db):
    user = User.objects.create_user("testuser", "test@example.com", "password")
    recipe = Recipe.objects.create(
        title="Inflation Recipe", price=5.00, uploaded_by=user
    )

    order = Order.objects.create(user=user)
    item = OrderItem.objects.create(order=order, recipe=recipe)
    order.total_price = item.price
    order.save()

    assert order.total_price == 5.00
    assert item.price == 5.00

    # Change recipe price
    recipe.price = 20.00
    recipe.save()

    # Order item price should remain the same
    item.refresh_from_db()
    assert item.price == 5.00
