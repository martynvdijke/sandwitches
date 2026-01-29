import pytest
import re
from playwright.sync_api import Page, expect
from sandwitches.models import User, Recipe, Order


@pytest.fixture
def staff_user(db):
    return User.objects.create_user(
        username="staff_ui",
        email="staff@example.com",
        password="password",
        is_staff=True,
    )


@pytest.mark.django_db
def test_admin_orders_page(page: Page, live_server, staff_user):
    # Ensure superuser exists so we don't hit setup
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Create some data
    recipe = Recipe.objects.create(  # ty:ignore[unresolved-attribute]
        title="UI Test Recipe", price=12.50, uploaded_by=staff_user
    )
    user = User.objects.create_user("customer", "cust@example.com", "password")
    Order.objects.create(user=user, recipe=recipe)  # ty:ignore[unresolved-attribute]

    # Login as staff
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Wait for login to complete
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to Dashboard
    page.goto(f"{live_server.url}/dashboard/")

    # Navigate to Orders via URL or Menu?
    # Let's try Menu if visible, or just URL for robustness in this test
    page.goto(f"{live_server.url}/dashboard/orders/")

    # Check page title/header
    expect(page.get_by_role("heading", name="Orders", level=5)).to_be_visible()

    # Check table content
    # Look for the recipe title and user
    expect(page.get_by_role("cell", name="UI Test Recipe")).to_be_visible()
    expect(page.get_by_text("customer")).to_be_visible()
    expect(page.get_by_text("€ 12.50")).to_be_visible()
    # TODO: Check status - pending by default
    # expect(page.get_by_text("Pending")).to_be_visible()


@pytest.mark.django_db
def test_admin_recipe_management_new_fields(page: Page, live_server, staff_user):
    # Ensure superuser
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Wait for login to complete
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to Add Recipe
    page.goto(f"{live_server.url}/dashboard/recipes/add/")

    # Fill standard fields
    page.fill("input[name='title']", "New Deluxe Sandwich")
    page.fill("input[name='tags_string']", "deluxe, tasty")

    # The description field uses EasyMDE which hides the textarea.
    # Playwright might struggle to fill it directly if it's hidden.
    # But usually we can fill the underlying textarea or use type into the editor.
    # For now, let's try filling the textarea if visible, or skip if complex (EasyMDE acts as a div).
    # Since existing tests didn't test admin add, let's assume we might need to handle this.
    # For simplicity, we can skip description/ingredients/instructions or try to set them.
    # Let's try setting them via JS if needed, or just type if standard.
    # Wait, the template uses EasyMDE.

    # Let's focus on the NEW fields: Price, Max Daily Orders, Highlight
    page.fill("input[name='price']", "25.99")
    page.fill("input[name='max_daily_orders']", "50")

    # Highlight is a checkbox
    # BeerCSS/Material checkbox structure might hide the input.
    # Usually <label class="checkbox"><input type="checkbox"><span>Label</span></label>
    # So we can check by label text or locator
    page.check("text=Highlighted")

    # Save
    page.click("button:has-text('Save Recipe')")

    # Expect redirect to list
    expect(page).to_have_url(re.compile(r".*/dashboard/recipes/$"))

    # Verify in list
    # Row should contain title, price, orders info, and highlight icon
    row = page.get_by_role("row", name="New Deluxe Sandwich")
    expect(row).to_be_visible()

    # Check Price column
    expect(row).to_contain_text("€ 25.99")

    # Check Orders column (0 / 50)
    expect(row).to_contain_text("0 / 50")

    # Check Highlight (icon name is 'star' and class 'amber-text')
    # Use a more specific locator inside the row
    expect(row.locator("i.amber-text")).to_have_text("star")


@pytest.mark.django_db
def test_admin_tag_management_ui(page: Page, live_server, staff_user):
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to Tags
    page.goto(f"{live_server.url}/dashboard/tags/")

    # Add Tag
    page.click("a:has-text('Add Tag')")
    page.fill("input[name='name']", "Spicy")
    page.click("button:has-text('Save')")  # Assuming button text is Save

    expect(page).to_have_url(re.compile(r".*/dashboard/tags/$"))
    # Use exact match or role to avoid ambiguity with breadcrumbs/headers
    expect(page.get_by_role("cell", name="Spicy", exact=True)).to_be_visible()

    # Edit Tag
    page.click("tr:has-text('Spicy') td >> nth=0")
    page.fill("input[name='name']", "Extra Spicy")
    page.click("button:has-text('Save')")
    expect(page.get_by_role("cell", name="Extra Spicy", exact=True)).to_be_visible()

    # Delete Tag
    # Target the delete button in the row for "Extra Spicy"
    page.locator("tr:has-text('Extra Spicy')").get_by_role(
        "link", name="delete"
    ).click()
    page.click("button:has-text('Delete')")  # Confirm delete
    expect(page.get_by_text("Tag deleted.").first).to_be_visible()
    expect(page.get_by_text("Extra Spicy")).not_to_be_visible()


@pytest.mark.django_db
def test_admin_user_management_ui(page: Page, live_server, staff_user):
    User.objects.create_superuser("admin", "admin@example.com", "password")
    other_user = User.objects.create_user("to_edit", "edit@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to Users
    page.goto(f"{live_server.url}/dashboard/users/")

    # Edit User
    page.click("tr:has-text('to_edit')")
    page.check("text=Staff status")
    page.click("button:has-text('Save')")

    expect(page.get_by_text("User updated successfully.").first).to_be_visible()

    other_user.refresh_from_db()
    assert other_user.is_staff is True


@pytest.mark.django_db
def test_admin_rating_management_ui(page: Page, live_server, staff_user):
    User.objects.create_superuser("admin", "admin@example.com", "password")
    recipe = Recipe.objects.create(
        title="Rating Test", price=10, uploaded_by=staff_user
    )
    from sandwitches.models import Rating

    Rating.objects.create(
        recipe=recipe, user=staff_user, score=1.0, comment="Terrible!"
    )

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to Ratings
    page.goto(f"{live_server.url}/dashboard/ratings/")

    expect(page.locator("tr:has-text('Rating Test')")).to_contain_text("Terrible!")

    # Delete Rating
    page.click("a[title='Delete']")
    page.click("button:has-text('Delete')")

    expect(page.get_by_text("Rating deleted.").first).to_be_visible()
    expect(page.get_by_text("Terrible!")).not_to_be_visible()


@pytest.mark.django_db
def test_admin_gotify_settings_ui(page: Page, live_server):
    # Use superuser directly to avoid PermissionDenied in Django Admin
    User.objects.create_superuser("admin_ui", "admin@ui.com", "password")

    # Ensure Setting instance exists
    from sandwitches.models import Setting

    Setting.objects.get_or_create()

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "admin_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Go to Site Settings (Django Admin via Solo)
    # Solo redirects but if it fails in test environment, we go direct
    setting = Setting.get_solo()
    page.goto(f"{live_server.url}/admin/sandwitches/setting/{setting.pk}/change/")

    # Fill Gotify fields
    page.fill("input[name='gotify_url']", "https://gotify.test.com")
    page.fill("input[name='gotify_token']", "test-token-123")

    # Save
    page.click("input[name='_save']")

    # Verify persistence
    from sandwitches.models import Setting

    config = Setting.get_solo()
    assert config.gotify_url == "https://gotify.test.com"
    assert config.gotify_token == "test-token-123"


@pytest.mark.django_db
@pytest.mark.skip(reason="Flaky test - needs investigation")
def test_admin_photo_rotation_ui(page: Page, live_server, staff_user):
    # Ensure superuser
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Create recipe with VALID image (1x1 transparent GIF)
    from django.core.files.uploadedfile import SimpleUploadedFile

    gif_data = (
        b"\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff"
        b"\x00\x00\x00\x21\xf9\x04\x01\x00\x00\x00\x00\x2c\x00\x00\x00\x00"
        b"\x01\x00\x01\x00\x00\x02\x02\x44\x01\x00\x3b"
    )
    image = SimpleUploadedFile("test.gif", gif_data, content_type="image/gif")
    recipe = Recipe.objects.create(
        title="Rotate Me", image=image, uploaded_by=staff_user
    )

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "staff_ui")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Go to edit recipe
    page.goto(f"{live_server.url}/dashboard/recipes/{recipe.pk}/edit/")

    # Click "Edit" button to open cropper
    page.click("button:has-text('Edit')")

    # Verify cropper dialog is visible
    expect(page.locator("#cropper-dialog")).to_be_visible()

    # Find rotation buttons (rotate_left, rotate_right icons)
    expect(page.locator("i:has-text('rotate_left')")).to_be_visible()
    expect(page.locator("i:has-text('rotate_right')")).to_be_visible()

    # Click rotate right
    page.click("i:has-text('rotate_right')")

    # Click Apply
    page.click("button:has-text('Apply')")

    # Dialog should close
    expect(page.locator("#cropper-dialog")).not_to_be_visible()

    # Check that image_data hidden field is now populated with base64
    expect(page.locator("input[name='image_data']")).not_to_have_value("")
