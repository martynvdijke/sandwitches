import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe, Tag

User = get_user_model()


@pytest.mark.django_db
def test_admin_dashboard_requires_staff(client):
    url = reverse("admin_dashboard")
    response = client.get(url)
    # staff_member_required redirects to login
    assert response.status_code == 302
    assert "login" in response.url


@pytest.mark.django_db
def test_admin_dashboard_accessible_by_staff(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")
    response = client.get(url)
    assert response.status_code == 200
    assert "Dashboard" in response.content.decode()


@pytest.mark.django_db
def test_admin_recipe_list(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)
    Recipe.objects.create(title="Admin Recipe", description="Desc")

    url = reverse("admin_recipe_list")
    response = client.get(url)
    assert response.status_code == 200
    assert "Admin Recipe" in response.content.decode()


@pytest.mark.django_db
def test_admin_user_list(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)

    url = reverse("admin_user_list")
    response = client.get(url)
    assert response.status_code == 200
    assert "staff" in response.content.decode()


@pytest.mark.django_db
def test_admin_recipe_add_with_tags(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)

    url = reverse("admin_recipe_add")
    data = {
        "title": "New Tagged Recipe",
        "description": "Desc",
        "ingredients": "Ing",
        "instructions": "Inst",
        "tags_string": "tag-a, tag-b",
    }
    response = client.post(url, data, follow=True)
    assert response.status_code == 200

    recipe = Recipe.objects.get(title="New Tagged Recipe")
    assert recipe.tags.filter(name="tag-a").exists()
    assert recipe.tags.filter(name="tag-b").exists()


@pytest.mark.django_db
def test_admin_tag_management(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)

    # Add
    url_add = reverse("admin_tag_add")
    client.post(url_add, {"name": "TestTag"})
    tag = Tag.objects.get(name="TestTag")

    # List
    url_list = reverse("admin_tag_list")
    resp = client.get(url_list)
    assert "TestTag" in resp.content.decode()

    # Edit
    url_edit = reverse("admin_tag_edit", kwargs={"pk": tag.pk})
    client.post(url_edit, {"name": "UpdatedTag"})
    tag.refresh_from_db()
    assert tag.name == "UpdatedTag"

    # Delete
    url_del = reverse("admin_tag_delete", kwargs={"pk": tag.pk})
    client.post(url_del)
    assert not Tag.objects.filter(name="UpdatedTag").exists()


@pytest.mark.django_db
def test_admin_task_list_accessible(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_task_list")
    resp = client.get(url)
    assert resp.status_code == 200
    assert "Task Results" in resp.content.decode()
