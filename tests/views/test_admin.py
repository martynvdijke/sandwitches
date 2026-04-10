import pytest
from django.contrib.auth import get_user_model
from django.contrib.auth.models import Group
from django.urls import reverse

from sandwitches.models import Recipe, Tag

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
def test_admin_dashboard_date_range_and_gaps(client):
    staff = User.objects.create_user(
        username="staff_date", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")

    # Create a recipe today and one 5 days ago
    from datetime import timedelta

    from django.utils import timezone

    today = timezone.now()
    five_days_ago = today - timedelta(days=5)

    # We need to use a mock or manually set created_at if possible,
    # but created_at has auto_now_add=True.
    # We can use freezegun or just create them and they will be "today".
    Recipe.objects.create(title="Today Recipe", description="Desc")

    # To test historical data, we might need to update the created_at field after creation
    r_old = Recipe.objects.create(title="Old Recipe", description="Desc")
    Recipe.objects.filter(pk=r_old.pk).update(created_at=five_days_ago)

    # Request dashboard with a range covering both
    start_date = (five_days_ago - timedelta(days=1)).strftime("%Y-%m-%d")
    end_date = today.strftime("%Y-%m-%d")

    response = client.get(url, {"start_date": start_date, "end_date": end_date})
    assert response.status_code == 200

    recipe_counts = response.context["recipe_counts"]
    recipe_labels = response.context["recipe_labels"]

    # Should have at least 7 days (start to end inclusive)
    assert len(recipe_labels) >= 7
    assert len(recipe_counts) == len(recipe_labels)

    # One recipe 5 days ago, one today. The rest should be 0.
    assert sum(recipe_counts) == 2

    # Check if start_date and end_date in context are correct
    assert response.context["start_date"] == start_date
    assert response.context["end_date"] == end_date


@pytest.mark.django_db
def test_admin_dashboard_ratings_avg(client):
    staff = User.objects.create_user(
        username="staff_rating", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")

    # Create a recipe and ratings
    r = Recipe.objects.create(title="Rating Recipe", description="Desc")
    from datetime import timedelta

    from django.utils import timezone

    from sandwitches.models import Rating

    today = timezone.now()

    Rating.objects.create(recipe=r, user=staff, score=8.0)

    # Rating yesterday
    yesterday = today - timedelta(days=1)
    r2 = Rating.objects.create(
        recipe=r, user=User.objects.create_user("u2", "u2@e.c", "pw"), score=6.0
    )
    Rating.objects.filter(pk=r2.pk).update(created_at=yesterday)

    response = client.get(url)
    assert response.status_code == 200

    rating_labels = response.context["rating_labels"]
    rating_avgs = response.context["rating_avgs"]

    # We expect 2 ratings in the context if they fall in the default 30-day range
    assert len(rating_labels) == 2
    assert 8.0 in rating_avgs
    assert 6.0 in rating_avgs


@pytest.mark.django_db
def test_admin_dashboard_pending_recipes(client, community_group):
    staff = User.objects.create_user(
        username="staff_pending", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")

    # Create a community user
    comm_user = User.objects.create_user(username="comm_user", password="password")
    comm_user.groups.add(community_group)

    # Create a pending recipe
    Recipe.objects.create(
        title="Pending Recipe",
        description="Desc",
        uploaded_by=comm_user,
        is_approved=False,
    )

    # Create an approved recipe (should not be in pending_recipes)
    Recipe.objects.create(
        title="Approved Recipe",
        description="Desc",
        uploaded_by=comm_user,
        is_approved=True,
    )

    # Create a staff recipe (should not be in pending_recipes)
    Recipe.objects.create(
        title="Staff Recipe", description="Desc", uploaded_by=staff, is_approved=False
    )

    response = client.get(url)
    assert response.status_code == 200

    pending_recipes = response.context["pending_recipes"]
    assert len(pending_recipes) == 1
    assert pending_recipes[0].title == "Pending Recipe"
    assert "Pending Recipe" in response.content.decode()
    assert (
        "Approved Recipe" in response.content.decode()
    )  # Recent recipes might show it?
    # Actually, recent_recipes shows last 5.
    # Let's check if "Approved Recipe" is NOT in the pending list in the content.
    # We can check for the text "Pending Approvals" and then what's below it.


@pytest.mark.django_db
def test_admin_dashboard_empty_state(client):
    staff = User.objects.create_user(
        username="staff_empty", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_dashboard")

    # Ensure no data
    Recipe.objects.all().delete()
    Tag.objects.all().delete()
    User.objects.exclude(pk=staff.pk).delete()
    from sandwitches.models import Order, Rating

    Order.objects.all().delete()
    Rating.objects.all().delete()

    response = client.get(url)
    assert response.status_code == 200

    assert response.context["recipe_count"] == 0
    assert response.context["user_count"] == 1  # Just the staff user
    assert response.context["tag_count"] == 0
    assert (
        len(response.context["recipe_labels"]) > 0
    )  # Should still have labels for the date range
    assert all(count == 0 for count in response.context["recipe_counts"])
    assert all(count == 0 for count in response.context["order_counts"])
    assert len(response.context["rating_labels"]) == 0


@pytest.mark.django_db
def test_admin_recipe_approval_list(client, community_group):
    staff = User.objects.create_user(
        username="staff_appr", password="password", is_staff=True
    )
    client.force_login(staff)
    url = reverse("admin_recipe_approval_list")

    # Create community user
    comm_user = User.objects.create_user(username="comm_appr", password="password")
    comm_user.groups.add(community_group)

    # Pending
    Recipe.objects.create(title="Approve Me", uploaded_by=comm_user, is_approved=False)
    # Approved
    Recipe.objects.create(title="Already Done", uploaded_by=comm_user, is_approved=True)

    response = client.get(url)
    assert response.status_code == 200
    assert "Approve Me" in response.content.decode()
    assert "Already Done" not in response.content.decode()


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
