import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from django.conf import settings

User = get_user_model()


@pytest.mark.django_db
def test_user_settings_post_language(client):
    user = User.objects.create_user(
        username="testuser", password="password", language="en"
    )
    client.login(username="testuser", password="password")

    url = reverse("user_settings")
    response = client.post(url, {"language": "nl", "theme": "light"}, follow=True)

    if response.status_code != 200 or user.language != "nl":
        # Refresh user from DB to check if it was actually updated before checking form errors
        user.refresh_from_db()
        if user.language != "nl":
            print(
                f"Form errors: {response.context.get('form').errors if response.context and response.context.get('form') else 'No form in context'}"
            )

    assert response.status_code == 200
    user.refresh_from_db()
    assert user.language == "nl"
    assert client.cookies[settings.LANGUAGE_COOKIE_NAME].value == "nl"
