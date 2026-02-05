import pytest
from unittest.mock import patch, MagicMock, ANY
from sandwitches.tasks import send_gotify_notification
from sandwitches.models import Setting, Recipe, User
from django.urls import reverse


@pytest.mark.django_db
def test_send_gotify_notification_not_configured():
    # Ensure not configured
    config = Setting.get_solo()
    config.gotify_url = ""
    config.gotify_token = ""
    config.save()

    with patch("requests.post") as mock_post:
        result = send_gotify_notification.func(title="Test", message="Hello")
        assert result is False
        mock_post.assert_not_called()


@pytest.mark.django_db
def test_send_gotify_notification_success():
    config = Setting.get_solo()
    config.gotify_url = "https://gotify.example.com"
    config.gotify_token = "mytoken"
    config.save()

    with patch("requests.post") as mock_post:
        mock_response = MagicMock()
        mock_response.raise_for_status.return_value = None
        mock_post.return_value = mock_response

        result = send_gotify_notification.func(
            title="Test", message="Hello", priority=7
        )
        assert result is True

        mock_post.assert_called_once()
        args, kwargs = mock_post.call_args
        assert "https://gotify.example.com/message?token=mytoken" in args[0]
        assert kwargs["json"]["title"] == "Test"
        assert kwargs["json"]["message"] == "Hello"
        assert kwargs["json"]["priority"] == 7


@pytest.mark.django_db
def test_recipe_save_enqueues_gotify():
    with patch("sandwitches.models.send_gotify_notification") as mock_task:
        Recipe.objects.create(title="New Recipe", price=10)
        mock_task.enqueue.assert_called_with(
            title="New Recipe Uploaded",
            message="A new recipe 'New Recipe' has been uploaded by Unknown.",
            priority=5,
        )


@pytest.mark.django_db
def test_order_recipe_view_enqueues_gotify(client):
    User.objects.create_user(username="buyer", password="pw")
    client.login(username="buyer", password="pw")
    recipe = Recipe.objects.create(title="Order Me", price=15)

    with patch("sandwitches.views.send_gotify_notification") as mock_task:
        client.post(reverse("order_recipe", args=[recipe.pk]))
        mock_task.enqueue.assert_called_with(
            title="New Order Received",
            message=ANY,
            priority=6,
        )


@pytest.mark.django_db
def test_checkout_cart_view_enqueues_gotify(client):
    user = User.objects.create_user(username="cartbuyer", password="pw")
    client.login(username="cartbuyer", password="pw")
    recipe = Recipe.objects.create(title="Cart Item", price=10)
    from sandwitches.models import CartItem

    CartItem.objects.create(user=user, recipe=recipe, quantity=2)

    with patch("sandwitches.views.send_gotify_notification") as mock_task:
        client.post(reverse("checkout_cart"))
        mock_task.enqueue.assert_called_with(
            title="New Order Received",
            message=ANY,
            priority=6,
        )


@pytest.mark.django_db
def test_user_creation_enqueues_gotify():
    with patch("sandwitches.models.send_gotify_notification") as mock_task:
        User.objects.create_user(username="newuser", password="pw")
        mock_task.enqueue.assert_called_with(
            title="New User Created",
            message="User newuser has joined Sandwitches!",
            priority=4,
        )
