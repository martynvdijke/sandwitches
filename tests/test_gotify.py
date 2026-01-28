import pytest
from unittest.mock import patch, MagicMock, ANY
from sandwitches.tasks import send_gotify_notification
from sandwitches.models import Setting, Recipe, Order, User


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
def test_order_save_enqueues_gotify():
    user = User.objects.create_user(username="buyer", password="pw")
    recipe = Recipe.objects.create(title="Order Me", price=15)

    with patch("sandwitches.models.send_gotify_notification") as mock_task:
        Order.objects.create(user=user, recipe=recipe)
        mock_task.enqueue.assert_called_with(
            title="New Order Received",
            message=ANY,  # contains price and order ID which might be dynamic
            priority=6,
        )

        call_args = mock_task.enqueue.call_args
        message = call_args[1]["message"]
        assert "Order #" in message
        assert "Order Me" in message
        assert "buyer" in message


@pytest.mark.django_db
def test_user_creation_enqueues_gotify():
    with patch("sandwitches.models.send_gotify_notification") as mock_task:
        User.objects.create_user(username="newuser", password="pw")
        mock_task.enqueue.assert_called_with(
            title="New User Created",
            message="User newuser has joined Sandwitches!",
            priority=4,
        )
