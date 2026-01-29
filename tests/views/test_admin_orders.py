import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe, Order

User = get_user_model()


@pytest.mark.django_db
def test_admin_order_list_view(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    user = User.objects.create_user(username="user", password="password")
    recipe = Recipe.objects.create(title="Recipe 1", price=10.00, uploaded_by=staff)
    Order.objects.create(user=user, recipe=recipe)

    client.force_login(staff)
    url = reverse("admin_order_list")

    # Test standard GET request
    response = client.get(url)
    assert response.status_code == 200
    assert "Orders" in response.content.decode()
    assert "Recipe 1" in response.content.decode()


@pytest.mark.django_db
def test_admin_order_list_htmx(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    user = User.objects.create_user(username="user", password="password")
    recipe = Recipe.objects.create(title="Recipe 1", price=10.00, uploaded_by=staff)
    Order.objects.create(user=user, recipe=recipe)

    client.force_login(staff)
    url = reverse("admin_order_list")

    # Test HTMX request
    response = client.get(url, headers={"HX-Request": "true"})
    assert response.status_code == 200
    content = response.content.decode()
    # Should render the partial, so it should contain the row but NOT the full page header
    assert "Recipe 1" in content
    assert (
        "Orders" not in content
    )  # "Orders" is in the block admin_title/header of the full page, not partial row


@pytest.mark.django_db
def test_admin_recipe_form_new_fields(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)

    url = reverse("admin_recipe_add")
    response = client.get(url)
    assert response.status_code == 200
    content = response.content.decode()

    # Check for presence of new fields
    assert 'name="price"' in content
    assert 'name="is_highlighted"' in content
    assert 'name="max_daily_orders"' in content


@pytest.mark.django_db
def test_admin_recipe_list_columns(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)
    Recipe.objects.create(
        title="List Test Recipe",
        price=15.50,
        daily_orders_count=5,
        max_daily_orders=10,
        is_highlighted=True,
    )

    url = reverse("admin_recipe_list")
    response = client.get(url)
    assert response.status_code == 200
    content = response.content.decode()

    # Check for new columns/values
    assert "€ 15.50" in content
    assert "5 / 10" in content
    assert "star" in content  # Highlight icon
