import pytest
from unittest.mock import patch, MagicMock
from sandwitches.tasks import email_users, send_emails
from sandwitches.models import Recipe
from django.contrib.auth import get_user_model
from django.core.files.uploadedfile import SimpleUploadedFile

User = get_user_model()


@pytest.mark.django_db
def test_email_users_task_no_users():
    # If no users with email, should return 0
    context = MagicMock()
    context.attempt = 1
    context.task_result.id = "123"

    # Ensure no users
    User.objects.all().delete()

    with patch("sandwitches.tasks.send_emails") as mock_send:
        # call the underlying function
        result = email_users.func(context=context, recipe_id=1)
        assert result == 0
        mock_send.assert_not_called()


@pytest.mark.django_db
def test_email_users_task_with_users():
    context = MagicMock()
    context.attempt = 1
    context.task_result.id = "123"

    User.objects.create_user("u1", "u1@example.com", "p1")
    User.objects.create_user("u2", "u2@example.com", "p2")
    User.objects.create_user("u3", "", "p3")  # No email

    with patch("sandwitches.tasks.send_emails") as mock_send:
        result = email_users.func(context=context, recipe_id=99)
        assert result is True

        args, _ = mock_send.call_args
        assert args[0] == 99
        emails = args[1]
        assert len(emails) == 2
        assert "u1@example.com" in emails
        assert "u2@example.com" in emails


@pytest.mark.django_db
def test_send_emails(settings):
    # Setup
    settings.EMAIL_FROM_ADDRESS = "noreply@example.com"

    image = SimpleUploadedFile("test_image.jpg", b"content", content_type="image/jpeg")
    r = Recipe.objects.create(title="Yummy", description="So good", image=image)
    emails = ["a@a.com", "b@b.com"]

    with patch("sandwitches.tasks.EmailMultiAlternatives") as MockEmail:
        mock_msg = MagicMock()
        MockEmail.return_value = mock_msg

        send_emails(r.pk, emails)

        call_kwargs = MockEmail.call_args[1]

        assert call_kwargs["from_email"] == "noreply@example.com"
        assert "Yummy" in call_kwargs["subject"]
        assert "So good" in call_kwargs["body"]
        assert "html" in mock_msg.attach_alternative.call_args[0][1]
