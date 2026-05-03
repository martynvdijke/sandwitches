import pytest
from django.contrib.auth import get_user_model
from django.urls import reverse

User = get_user_model()


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
