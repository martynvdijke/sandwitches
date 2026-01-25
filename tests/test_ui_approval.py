import pytest
import re
from playwright.sync_api import Page, expect
from sandwitches.models import User, Recipe
from django.contrib.auth.models import Group


@pytest.fixture
def staff_user(db):
    user = User.objects.create_user(
        username="admin_ui",
        email="admin@example.com",
        password="password",
        is_staff=True,
        is_superuser=True,
    )
    admin_group, _ = Group.objects.get_or_create(name="admin")
    user.groups.add(admin_group)
    return user


@pytest.fixture
def regular_user(db):
    user = User.objects.create_user(
        username="user_ui", email="user@example.com", password="password"
    )
    community_group, _ = Group.objects.get_or_create(name="community")
    user.groups.add(community_group)
    return user


@pytest.mark.django_db
def test_recipe_submission_and_approval_flow(
    page: Page, live_server, staff_user, regular_user
):
    # 1. Login as regular user
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "user_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # 2. Go to Community
    # Open sidebar first
    page.click("button:has-text('menu')")
    page.click("a:has-text('Community')")

    # 3. Fill submission form
    page.fill("input[name='title']", "Community Sandwich")
    page.fill("textarea[name='description']", "Made by the community.")
    page.fill("textarea[name='ingredients']", "Love\nPeace")
    page.fill("textarea[name='instructions']", "Mix both.")
    page.fill("input[name='price']", "0.00")

    page.click("button:has-text('Submit Recipe')")

    # 4. Verify success message and redirect to profile
    expect(page).to_have_url(re.compile(r".*/profile/$"))
    expect(page.locator("body")).to_contain_text(
        "Your recipe has been submitted and is awaiting admin approval."
    )

    # 5. Verify it's NOT on the main index
    recipe = Recipe.objects.get(title="Community Sandwich")
    assert recipe.uploaded_by.groups.filter(name="community").exists()
    assert recipe.is_approved is False

    # Logout and check index as anonymous
    page.context.clear_cookies()

    # Let's just go back to index.
    page.goto(f"{live_server.url}/")
    expect(page.get_by_text("Community Sandwich")).not_to_be_visible()

    # 6. Login as Admin and approve
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "admin_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Go to the new Approvals page
    page.goto(f"{live_server.url}/dashboard/approvals/")

    row = page.get_by_role("row", name="Community Sandwich")
    expect(row).to_be_visible()

    # Click Approve
    row.get_by_role("link", name="check").click()

    # Verify it's approved
    expect(
        page.get_by_text("Recipe 'Community Sandwich' approved.").first
    ).to_be_visible()

    # Should be gone from approvals page table row
    expect(page.get_by_role("row", name="Community Sandwich")).not_to_be_visible()

    recipe.refresh_from_db()
    assert recipe.is_approved is True

    # 7. Verify it's now on the community index
    page.goto(f"{live_server.url}/community/")
    expect(page.get_by_text("Community Sandwich")).to_be_visible()
