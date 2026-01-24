import pytest
import re
from playwright.sync_api import Page, expect
from sandwitches.models import User, Recipe


@pytest.fixture
def staff_user(db):
    return User.objects.create_user(
        username="admin_ui",
        email="admin@example.com",
        password="password",
        is_staff=True,
        is_superuser=True,
    )


@pytest.fixture
def regular_user(db):
    return User.objects.create_user(
        username="user_ui", email="user@example.com", password="password"
    )


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

    # 5. Verify it's NOT on the main index for anonymous users (or just check the approval flag in DB first)
    recipe = Recipe.objects.get(title="Community Sandwich")
    assert recipe.is_community_made is True

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

    page.goto(f"{live_server.url}/dashboard/")

    # Check Pending Approvals section
    pending_section = page.locator(".s12:has(h5:has-text('Pending Approvals'))")
    expect(pending_section).to_be_visible()

    row = pending_section.get_by_role("row", name="Community Sandwich")
    expect(row).to_be_visible()

    # Click Approve
    # Use the check icon button in that row
    row.get_by_role("link", name="check").click()

    # Verify it's approved
    expect(
        page.get_by_text("Recipe 'Community Sandwich' approved.").first
    ).to_be_visible()
    expect(page.get_by_text("Pending Approvals")).not_to_be_visible()

    recipe.refresh_from_db()
    assert recipe.is_community_made is True

    # 7. Verify it's now on the community index
    page.goto(f"{live_server.url}/community/")
    expect(page.get_by_text("Community Sandwich")).to_be_visible()
