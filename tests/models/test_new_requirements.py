import pytest
from django.core.exceptions import ValidationError
from sandwitches.models import User, Order, Recipe
from decimal import Decimal


@pytest.mark.django_db
def test_user_email_validation():
    # Test valid email
    user = User(username="validuser", email="test@example.com")
    user.set_password("password123")
    user.full_clean()  # Should not raise
    user.save()

    # Test invalid email
    user_invalid = User(username="invaliduser", email="not-an-email")
    user_invalid.set_password("password123")
    with pytest.raises(ValidationError) as excinfo:
        user_invalid.full_clean()
    assert "email" in excinfo.value.message_dict


@pytest.mark.django_db
def test_order_status_and_completed():
    user = User.objects.create_user(
        username="orderuser", email="order@example.com", password="password"
    )
    recipe = Recipe.objects.create(
        title="Test Sandwich", price=Decimal("10.00"), servings=1
    )

    # Test new status choices
    order = Order.objects.create(
        user=user, recipe=recipe, status="PREPARING", completed=False
    )

    assert order.status == "PREPARING"
    assert not order.completed

    # Update status and completed
    order.status = "SHIPPED"
    order.completed = True
    order.save()

    order.refresh_from_db()
    assert order.status == "SHIPPED"
    assert order.completed


@pytest.mark.django_db
def test_unique_username():
    User.objects.create_user(
        username="uniqueuser", email="u1@example.com", password="password"
    )

    user2 = User(username="uniqueuser", email="u2@example.com")
    user2.set_password("password")
    with pytest.raises(ValidationError):
        user2.full_clean()
