import pytest
from django.core import mail
from sandwitches.models import Recipe, Order, OrderItem
from sandwitches.tasks import reset_daily_orders, notify_order_submitted
from django_tasks.backends.database.models import DBTaskResult


@pytest.mark.django_db
def test_reset_daily_orders_task(user_factory):
    user = user_factory()
    recipe1 = Recipe.objects.create(
        title="Recipe 1", price=10.00, daily_orders_count=5, uploaded_by=user
    )
    recipe2 = Recipe.objects.create(
        title="Recipe 2", price=12.00, daily_orders_count=10, uploaded_by=user
    )

    count = reset_daily_orders.func()

    recipe1.refresh_from_db()
    recipe2.refresh_from_db()

    assert recipe1.daily_orders_count == 0
    assert recipe2.daily_orders_count == 0
    assert count == 2


@pytest.mark.django_db
def test_notify_order_submitted_task(user_factory):
    user = user_factory(email="test@example.com", first_name="Test", last_name="User")
    recipe = Recipe.objects.create(
        title="Delicious Sandwich", price=8.50, uploaded_by=user
    )
    order = Order.objects.create(user=user, total_price=8.50)
    OrderItem.objects.create(order=order, recipe=recipe, quantity=1)

    notify_order_submitted.func(order.id)

    assert len(mail.outbox) == 1
    email = mail.outbox[0]
    assert email.to == ["test@example.com"]
    # Subject changed to Order #ID
    assert f"Order Confirmation: Order #{order.id}" in email.subject
    assert "Order ID: " + str(order.id) in email.body
    assert "Delicious Sandwich" in email.body


@pytest.mark.django_db
def test_create_order_api(client, user_factory):
    user = user_factory(email="api_test@example.com")
    recipe = Recipe.objects.create(title="API Recipe", price=10.00, uploaded_by=user)
    client.force_login(user)

    response = client.post(
        "/api/v1/orders", {"recipe_id": recipe.id}, content_type="application/json"
    )

    assert response.status_code == 201
    data = response.json()
    assert data["status"] == "PENDING"
    assert float(data["total_price"]) == 10.00

    assert Order.objects.filter(user=user).exists()
    order = Order.objects.filter(user=user).last()
    assert order.items.filter(recipe=recipe).exists()

    # Check if task was enqueued
    # Since we use database backend, we can check DBTaskResult
    # The queue name is 'emails' for notify_order_submitted
    assert DBTaskResult.objects.filter(queue_name="emails").exists()
