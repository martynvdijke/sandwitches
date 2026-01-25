import pytest
from django.urls import reverse
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe, Tag
from django.contrib.auth.models import Group

User = get_user_model()


@pytest.fixture
def admin_group(db):
    group, _ = Group.objects.get_or_create(name="admin")
    return group


@pytest.fixture
def community_group(db):
    group, _ = Group.objects.get_or_create(name="community")
    return group


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
    # Check if orders chart labels are in content
    assert "order_labels" in response.context
    assert "order_counts" in response.context


@pytest.mark.django_db
def test_admin_dashboard_htmx_partial(client):
    staff = User.objects.create_user(
        username="staff_htmx", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")

    # Test HTMX request
    response = client.get(url, headers={"HX-Request": "true"})
    assert response.status_code == 200
    content = response.content.decode()

    # Should contain chart canvases but NOT the full page sidebar/nav
    assert "recipeChart" in content
    assert "orderChart" in content
    assert "ratingChart" in content
    assert "Dashboard" not in content  # "Dashboard" is in block admin_title


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


@pytest.mark.django_db
def test_admin_recipe_edit(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    recipe = Recipe.objects.create(title="Old Title", description="Old Desc")

    url = reverse("admin_recipe_edit", kwargs={"pk": recipe.pk})

    data = {
        "title": "New Title",
        "description": "New Desc",
        "ingredients": "Ing",
        "instructions": "Inst",
        "tags_string": "tag1",
        "uploaded_by": staff.pk,
    }

    resp = client.post(url, data, follow=True)

    assert resp.status_code == 200

    recipe.refresh_from_db()

    assert recipe.title == "New Title"

    assert recipe.uploaded_by == staff


@pytest.mark.django_db
def test_admin_recipe_delete(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    recipe = Recipe.objects.create(title="To Delete", description="Desc")

    url = reverse("admin_recipe_delete", kwargs={"pk": recipe.pk})

    # GET shows confirm page

    resp = client.get(url)

    assert resp.status_code == 200

    assert "Are you sure?" in resp.content.decode()

    # POST deletes

    client.post(url)

    assert not Recipe.objects.filter(pk=recipe.pk).exists()


@pytest.mark.django_db
def test_admin_user_edit(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    other_user = User.objects.create_user(username="other", password="password")

    url = reverse("admin_user_edit", kwargs={"pk": other_user.pk})

    data = {
        "username": "other_updated",
        "email": "other@example.com",
        "is_staff": True,
        "is_active": True,
        "language": "nl",
    }

    client.post(url, data)

    other_user.refresh_from_db()

    assert other_user.username == "other_updated"

    assert other_user.is_staff is True

    assert other_user.language == "nl"


@pytest.mark.django_db
def test_admin_user_delete(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    other_user = User.objects.create_user(username="other", password="password")

    url = reverse("admin_user_delete", kwargs={"pk": other_user.pk})

    client.post(url)

    assert not User.objects.filter(pk=other_user.pk).exists()


@pytest.mark.django_db
def test_admin_rating_management(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    recipe = Recipe.objects.create(title="Recipe", description="Desc")

    from sandwitches.models import Rating

    rating = Rating.objects.create(recipe=recipe, user=staff, score=8.5)

    # List

    url_list = reverse("admin_rating_list")

    resp = client.get(url_list)

    assert "8.5" in resp.content.decode()

    # Delete

    url_del = reverse("admin_rating_delete", kwargs={"pk": rating.pk})

    client.post(url_del)

    assert not Rating.objects.filter(pk=rating.pk).exists()


@pytest.mark.django_db
def test_admin_task_detail(client):
    staff = User.objects.create_user(
        username="staff", password="password", is_staff=True
    )

    client.force_login(staff)

    from django_tasks.backends.database.models import DBTaskResult

    task = DBTaskResult.objects.create(
        task_path="sandwitches.tasks.email_users",
        status="SUCCEEDED",
        queue_name="default",
        priority=0,
        args_kwargs="[]",
        backend_name="default",
    )

    url = reverse("admin_task_detail", kwargs={"pk": task.id})

    resp = client.get(url)

    assert resp.status_code == 200

    assert "sandwitches.tasks.email_users" in resp.content.decode()


def test_recipe_form_rotation_logic(db, monkeypatch):
    # Mock PIL Image
    class MockImage:
        def rotate(self, angle, expand=False):
            self.rotated_angle = angle
            return self

        def save(self, path):
            self.saved_path = path

    mock_img = MockImage()

    import PIL.Image

    monkeypatch.setattr(PIL.Image, "open", lambda p: mock_img)

    staff = User.objects.create_user(
        username="staff_rot", password="password", is_staff=True
    )
    recipe = Recipe.objects.create(title="Rotation Test", image="test.jpg")

    from sandwitches.forms import RecipeForm

    data = {
        "title": "Rotation Test",
        "rotation": 90,
        "description": "Desc",
        "ingredients": "Ing",
        "instructions": "Inst",
        "uploaded_by": staff.pk,
    }
    form = RecipeForm(data=data, instance=recipe)
    assert form.is_valid()
    form.save()

    # Check if rotate was called with -90 (PIL rotates CCW, we send CW)
    assert mock_img.rotated_angle == -90
    assert "test.jpg" in str(mock_img.saved_path)


@pytest.mark.django_db
def test_admin_user_delete_self_fails(client):
    staff = User.objects.create_user(
        username="staff_me", password="password", is_staff=True
    )
    client.force_login(staff)

    url = reverse("admin_user_delete", kwargs={"pk": staff.pk})
    resp = client.post(url, follow=True)

    # Should redirect and show error message
    assert resp.status_code == 200
    assert "You cannot delete yourself." in resp.content.decode()
    assert User.objects.filter(pk=staff.pk).exists()


@pytest.mark.django_db
def test_index_favorites_filter_isolation(client, admin_group):
    user1 = User.objects.create_user(username="u1", password="pw")
    admin = User.objects.create_superuser("admin_iso", "admin@example.com", "pw")
    admin.groups.add(admin_group)

    r1 = Recipe.objects.create(title="User1 Fav", description="D1", uploaded_by=admin)
    Recipe.objects.create(title="User2 Fav", description="D2", uploaded_by=admin)

    user1.favorites.add(r1)
    # r2 is not user1's favorite

    client.force_login(user1)
    resp = client.get("/", {"favorites": "on"})
    assert "User1 Fav" in resp.content.decode()
    assert "User2 Fav" not in resp.content.decode()


@pytest.mark.django_db
def test_admin_recipe_rotate_ccw(client, monkeypatch):
    class MockImage:
        def rotate(self, angle, expand=False):
            self.rotated_angle = angle
            return self

        def save(self, path):
            pass

    mock_img = MockImage()
    import PIL.Image

    monkeypatch.setattr(PIL.Image, "open", lambda p: mock_img)

    staff = User.objects.create_user(username="staff_ccw", password="pw", is_staff=True)
    client.force_login(staff)
    recipe = Recipe.objects.create(title="RotCCW", image="test.jpg")

    url = reverse("admin_recipe_rotate", kwargs={"pk": recipe.pk})
    client.get(url, {"direction": "ccw"})

    # CCW should call PIL rotate with +90 (since PIL is CCW positive)
    assert mock_img.rotated_angle == 90


@pytest.mark.django_db
def test_admin_task_detail_with_error(client):
    staff = User.objects.create_user(username="staff_err", password="pw", is_staff=True)
    client.force_login(staff)

    from django_tasks.backends.database.models import DBTaskResult

    task = DBTaskResult.objects.create(
        task_path="sandwitches.tasks.fail",
        status="FAILED",
        args_kwargs="[]",
        backend_name="default",
        exception_class_path="ValueError",
        traceback="Traceback details...",
    )

    url = reverse("admin_task_detail", kwargs={"pk": task.id})
    resp = client.get(url)
    assert "Traceback details..." in resp.content.decode()
    assert "ValueError" in resp.content.decode()


@pytest.mark.django_db
def test_index_template_attributes(client, admin_group):
    # Verify hx-swap="outerHTML" is present to prevent layout squishing
    admin = User.objects.create_superuser("admin_attr", "admin@example.com", "pw")
    admin.groups.add(admin_group)
    resp = client.get("/")
    assert 'hx-swap="outerHTML"' in resp.content.decode()
