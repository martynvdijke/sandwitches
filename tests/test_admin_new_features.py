import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe, Order, Setting
from sandwitches.utils import ORDER_DB

User = get_user_model()


@pytest.mark.django_db
def test_admin_settings_view(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_settings")

    # GET request
    response = client.get(url)
    assert response.status_code == 200
    assert "Site Settings" in response.content.decode()

    # POST request (update settings)
    data = {
        "site_name": "New Site Name",
        "site_description": "New Description",
        "email": "test@example.com",
    }
    response = client.post(url, data)
    assert response.status_code == 302  # Redirect after success
    assert response.url == url

    # Verify change
    setting = Setting.get_solo()
    assert setting.site_name == "New Site Name"


@pytest.mark.django_db
def test_admin_order_status_update(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    user = User.objects.create_user(username="user", password="password")
    recipe = Recipe.objects.create(title="Recipe 1", price=10.00)  # noqa: F841
    order = Order.objects.create(user=user, status="PENDING")

    client.force_login(staff)
    url = reverse("admin_order_update_status", kwargs={"pk": order.pk})

    # Update status to PREPARING
    response = client.post(url, {"status": "PREPARING"})
    assert response.status_code == 302
    order.refresh_from_db()
    assert order.status == "PREPARING"
    assert order.completed is False

    # Verify ORDER_DB update
    assert order.pk in ORDER_DB
    assert ORDER_DB[order.pk] == "PREPARING"

    # Update status to COMPLETED
    response = client.post(url, {"status": "COMPLETED"})
    assert response.status_code == 302
    order.refresh_from_db()
    assert order.status == "COMPLETED"
    assert order.completed is True

    # Verify ORDER_DB update
    assert ORDER_DB[order.pk] == "COMPLETED"


@pytest.mark.django_db
def test_order_status_immutability(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    user = User.objects.create_user(username="user", password="password")
    recipe = Recipe.objects.create(title="Recipe 1", price=10.00)  # noqa: F841

    # Completed order
    order_comp = Order.objects.create(user=user, status="COMPLETED", completed=True)
    client.force_login(staff)
    url_comp = reverse("admin_order_update_status", kwargs={"pk": order_comp.pk})

    response = client.post(url_comp, {"status": "PENDING"})
    assert response.status_code == 302
    order_comp.refresh_from_db()
    assert order_comp.status == "COMPLETED"  # Should not change

    # Cancelled order
    order_can = Order.objects.create(user=user, status="CANCELLED")
    url_can = reverse("admin_order_update_status", kwargs={"pk": order_can.pk})

    response = client.post(url_can, {"status": "PENDING"})
    assert response.status_code == 302
    order_can.refresh_from_db()
    assert order_can.status == "CANCELLED"  # Should not change
