import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Setting

User = get_user_model()


@pytest.mark.django_db
def test_instagram_enabled_auto_toggle():
    config = Setting.get_solo()

    # 1. Credentials provided -> should be enabled
    config.instagram_username = "user"
    config.instagram_password = "password"
    config.save()
    assert config.instagram_enabled is True

    # 2. Credentials removed -> should be disabled
    config.instagram_username = ""
    config.save()
    assert config.instagram_enabled is False

    # 3. Restored credentials -> re-enabled
    config.instagram_username = "user"
    config.save()
    assert config.instagram_enabled is True


@pytest.mark.django_db
def test_admin_logs_view_as_staff(client):
    User.objects.create_superuser(
        username="admin", password="password", email="admin@example.com"
    )
    client.login(username="admin", password="password")

    url = reverse("admin_logs")
    response = client.get(url)
    assert response.status_code == 200
    assert "System Logs" in response.content.decode()


@pytest.mark.django_db
def test_admin_logs_unauthorized(client):
    User.objects.create_user(username="user", password="password")
    client.login(username="user", password="password")

    url = reverse("admin_logs")
    response = client.get(url)
    # staff_member_required usually redirects to login if not staff
    assert response.status_code == 302
