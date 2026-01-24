import pytest
from django.urls import reverse
from sandwitches.models import Recipe, User

@pytest.mark.django_db
def test_recipe_separation(client):
    # Setup users
    admin = User.objects.create_superuser("admin", "admin@example.com", "pw")
    staff = User.objects.create_user("staff", "staff@example.com", "pw", is_staff=True)
    user = User.objects.create_user("user", "user@example.com", "pw")
    
    # Setup recipes
    r_admin = Recipe.objects.create(title="Admin Recipe", uploaded_by=admin, is_approved=True)
    r_staff = Recipe.objects.create(title="Staff Recipe", uploaded_by=staff, is_approved=True)
    r_system = Recipe.objects.create(title="System Recipe", uploaded_by=None, is_approved=True)
    r_user = Recipe.objects.create(title="User Recipe", uploaded_by=user, is_approved=True)
    r_user_pending = Recipe.objects.create(title="Pending User Recipe", uploaded_by=user, is_approved=False)

    # 1. Check Main Page (Index) - should show Admin, Staff, System. NOT User.
    # We need a superuser in DB for index to work (it checks for it)
    response = client.get(reverse("index"))
    content = response.content.decode()
    assert "Admin Recipe" in content
    assert "Staff Recipe" in content
    assert "System Recipe" in content
    assert "User Recipe" not in content
    assert "Pending User Recipe" not in content

    # 2. Check Community Page - should show User Recipe. NOT Admin, Staff, System.
    client.force_login(user)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    assert "Pending User Recipe" in content # User sees their own pending
    assert "Admin Recipe" not in content
    assert "Staff Recipe" not in content
    assert "System Recipe" not in content

    # 3. Check Community Page as another user
    other_user = User.objects.create_user("other", "other@example.com", "pw")
    client.force_login(other_user)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    assert "Pending User Recipe" not in content # Other user doesn't see pending

    # 4. Check Community Page as Staff
    client.force_login(staff)
    response = client.get(reverse("community"))
    content = response.content.decode()
    assert "User Recipe" in content
    # Staff/Admin see all community recipes? 
    # Current view logic: if not request.user.is_staff: recipes = recipes.filter(...)
    # So staff DOES see pending.
    assert "Pending User Recipe" in content
    assert "Admin Recipe" not in content
    assert "Staff Recipe" not in content
