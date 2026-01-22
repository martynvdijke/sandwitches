import logging
import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe

User = get_user_model()


@pytest.mark.django_db
def test_admin_recipe_sort(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)

    # Create recipes
    r1 = Recipe.objects.create(
        title="A Recipe", price=10, daily_orders_count=5, uploaded_by=staff
    )
    r2 = Recipe.objects.create(
        title="B Recipe", price=20, daily_orders_count=1, uploaded_by=staff
    )
    r3 = Recipe.objects.create(
        title="C Recipe", price=5, daily_orders_count=10, uploaded_by=staff
    )
    logging.debug(r1, r2, r3)

    url = reverse("admin_recipe_list")

    # Sort by Title (ASC)
    response = client.get(url, {"sort": "title"})
    content = response.content.decode()
    # Check order by finding indices
    idx1 = content.find("A Recipe")
    idx2 = content.find("B Recipe")
    idx3 = content.find("C Recipe")
    assert idx1 < idx2 < idx3

    # Sort by Title (DESC)
    response = client.get(url, {"sort": "-title"})
    content = response.content.decode()
    idx1 = content.find("A Recipe")
    idx2 = content.find("B Recipe")
    idx3 = content.find("C Recipe")
    assert idx3 < idx2 < idx1

    # Sort by Price (ASC) -> 5 (C), 10 (A), 20 (B)
    response = client.get(url, {"sort": "price"})
    content = response.content.decode()
    idxA = content.find("A Recipe")  # 10
    idxB = content.find("B Recipe")  # 20
    idxC = content.find("C Recipe")  # 5
    assert idxC < idxA < idxB

    # Sort by Orders (DESC) -> 10 (C), 5 (A), 1 (B)
    response = client.get(url, {"sort": "-orders"})
    content = response.content.decode()
    idxA = content.find("A Recipe")
    idxB = content.find("B Recipe")
    idxC = content.find("C Recipe")
    assert idxC < idxA < idxB
