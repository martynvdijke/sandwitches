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
    response = client.post(url, {"language": "nl"}, follow=True)

    assert response.status_code == 200
    user.refresh_from_db()
    assert user.language == "nl"
    assert client.cookies[settings.LANGUAGE_COOKIE_NAME].value == "nl"
