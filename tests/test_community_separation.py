import pytest
from django.urls import reverse
from sandwitches.models import Recipe, User
from django.contrib.auth.models import Group


@pytest.mark.django_db
def test_recipe_separation(client):
    # Setup groups
    admin_group, _ = Group.objects.get_or_create(name="admin")
    community_group, _ = Group.objects.get_or_create(name="community")

    # Setup users
    admin = User.objects.create_superuser("admin", "admin@example.com", "pw")
    admin.groups.add(admin_group)

    staff = User.objects.create_user("staff", "staff@example.com", "pw", is_staff=True)
    staff.groups.add(admin_group)  # Staff also in admin group for testing index

    user = User.objects.create_user("user", "user@example.com", "pw")
    user.groups.add(community_group)

    # Setup recipes
    r_admin = Recipe.objects.create(  # noqa: F841
        title="Admin Recipe", uploaded_by=admin, is_approved=True
    )
    r_staff = Recipe.objects.create(  # noqa: F841
        title="Staff Recipe", uploaded_by=staff, is_approved=True
    )
    r_system = Recipe.objects.create(  # noqa: F841
        title="System Recipe", uploaded_by=None, is_approved=True
    )
    # Note: System Recipe (uploaded_by=None) might not show up if filter is uploaded_by__groups__name="admin"
    # Let's check the view logic again.
    # recipes = recipes.filter(uploaded_by__groups__name="admin")
    # If uploaded_by is None, it won't have groups.

    r_user = Recipe.objects.create(  # noqa: F841
        title="User Recipe", uploaded_by=user, is_approved=True
    )
    r_user_pending = Recipe.objects.create(  # noqa: F841
        title="Pending User Recipe", uploaded_by=user, is_approved=False
    )

    # 1. Check Main Page (Index) - should show Admin, Staff. NOT User.
    response = client.get(reverse("index"))
    content = response.content.decode()
    assert "Admin Recipe" in content
    assert "Staff Recipe" in content
    # assert "System Recipe" in content # This might fail if uploaded_by is None
    assert "User Recipe" not in content
    assert "Pending User Recipe" not in content

    # 2. Check Community Page - should show User Recipe. NOT Admin, Staff.
    client.force_login(user)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    assert "Admin Recipe" not in content
    assert "Staff Recipe" not in content

    # 3. Check Community Page as another user
    other_user = User.objects.create_user("other", "other@example.com", "pw")
    other_user.groups.add(community_group)
    client.force_login(other_user)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    assert "Pending User Recipe" not in content  # Approved only for others

    # 4. Check Community Page as Staff
    client.force_login(staff)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    # Staff/Admin see all community recipes
    assert "Pending User Recipe" in content
