import pytest
import re
from playwright.sync_api import Page, expect
from sandwitches.models import User, Recipe, Order, OrderItem


@pytest.fixture
def user(db):
    return User.objects.create_user("testuser", "test@example.com", "password")


@pytest.fixture
def recipe(db, user):
    return Recipe.objects.create(
        title="Test Sandwich",
        description="Delicious test sandwich",
        uploaded_by=user,
        servings=1,
        ingredients="Bread\nCheese",
        instructions="Put cheese on bread.",
    )


@pytest.mark.django_db
def test_login_client(client, user):
    """
    Verify that login works at the Django level.
    """
    assert client.login(username="testuser", password="password")


@pytest.mark.django_db
def test_initial_setup_redirect(page: Page, live_server):
    """
    Test that the application redirects to the setup page when no admin exists.
    """
    # Ensure no users exist
    User.objects.all().delete()

    page.goto(live_server.url)

    # Should redirect to /setup/, potentially with language prefix (e.g. /en/setup/)
    expect(page).to_have_url(re.compile(r".*/setup/$"))
    expect(page).to_have_title("Initial setup — Create admin")

    heading = page.get_by_role("heading", name="Create initial administrator")
    expect(heading).to_be_visible()


@pytest.mark.django_db
def test_login_flow(page: Page, live_server, user):
    """
    Test the login flow and redirection to the index page.
    """
    # Create a superuser to ensure we don't get redirected to setup
    User.objects.create_superuser("admin", "admin@example.com", "password")

    page.goto(f"{live_server.url}/login/")

    # Fill login form using name attribute
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")

    # Submit via Enter key
    page.press("input[name='password']", "Enter")

    expect(page).to_have_url(f"{live_server.url}/")

    # Verify user is logged in by checking for the user menu avatar
    expect(page.locator("a[href*='/profile'] img")).to_be_visible()


@pytest.mark.django_db
def test_recipe_favoriting(page: Page, live_server, user, recipe):
    """
    Test favoriting a recipe.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    expect(page).to_have_url(f"{live_server.url}/")

    # Go to recipe detail
    page.goto(f"{live_server.url}/recipes/{recipe.slug}/")

    # Initial state: not favorited (icon is favorite_border)
    toggle_btn = page.locator("#favorite-toggle-button")
    expect(toggle_btn).to_be_visible()  # Ensure we are logged in
    expect(toggle_btn.locator("i")).to_have_text("favorite_border")

    # Click favorite
    toggle_btn.click()

    # Verify state change (icon is favorite)
    expect(toggle_btn.locator("i")).to_have_text("favorite")

    # Go to Favorites page
    page.goto(f"{live_server.url}/favorites/")

    # Verify recipe is listed
    expect(page.get_by_text(recipe.title)).to_be_visible()


@pytest.mark.django_db
def test_recipe_rating(page: Page, live_server, user, recipe):
    """
    Test rating and commenting on a recipe.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    expect(page).to_have_url(f"{live_server.url}/")

    page.goto(f"{live_server.url}/recipes/{recipe.slug}/")

    # Rate
    comment_text = "This is a fantastic test sandwich!"
    page.fill("textarea[name='comment']", comment_text)

    page.click("button:has-text('Submit Rating')")

    # Verify the rating appears on the page
    expect(page.get_by_text(comment_text).first).to_be_visible()
    expect(page.get_by_text("Your rating:")).to_be_visible()


@pytest.mark.django_db
def test_signup_flow(page: Page, live_server):
    """
    Test the user signup process.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")

    page.goto(f"{live_server.url}/signup/")

    new_username = "new_user_123"

    page.fill("input[name='first_name']", "New")
    page.fill("input[name='last_name']", "User")
    page.fill("input[name='username']", new_username)
    page.fill("input[name='email']", "new@example.com")
    strong_pass = "SecurePass123!"
    page.fill("input[name='password1']", strong_pass)
    page.fill("input[name='password2']", strong_pass)

    # Submit via Enter on last field
    page.press("input[name='password2']", "Enter")

    # Should redirect to index and be logged in
    expect(page).to_have_url(f"{live_server.url}/")
    expect(page.locator("a[href*='/profile'] img")).to_be_visible()


@pytest.mark.django_db
def test_order_sandwich_ui(page: Page, live_server, user, recipe):
    """
    Test ordering a sandwich and verifying the success message.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")
    recipe.price = 10.00
    recipe.save()

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to recipe
    page.goto(f"{live_server.url}/recipes/{recipe.slug}/")

    # Click Add to Cart
    page.click("button:has-text('Add to Cart')")

    # Verify success message (Added [title] to your cart.)
    expect(page.get_by_text(f"Added {recipe.title} to your cart.")).to_be_visible()


@pytest.mark.django_db
def test_scale_ingredients_ui(page: Page, live_server, recipe):
    """
    Test the ingredient scaling functionality.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")
    recipe.ingredients = "2 slices of Bread\n1 slice of Cheese"
    recipe.servings = 1
    recipe.is_approved = True
    recipe.save()

    page.goto(f"{live_server.url}/recipes/{recipe.slug}/")

    # Initial state (1 portion)
    display = page.locator("#ingredients-display")
    expect(display).to_contain_text("2 slices of Bread")

    # Increase to 2 portions
    page.click("button:has-text('+')")

    # Verify scaled ingredients
    expect(display).to_contain_text("4 slices of Bread")
    expect(display).to_contain_text("2")
    expect(display).to_contain_text("Cheese")


@pytest.mark.django_db
def test_update_profile_ui(page: Page, live_server, user):
    """
    Test updating user profile bio and names.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")
    expect(page).to_have_url(f"{live_server.url}/")

    # Go to profile
    page.goto(f"{live_server.url}/profile/")

    # Fill profile
    page.fill("input[name='first_name']", "UpdatedFirst")
    page.fill("input[name='last_name']", "UpdatedLast")
    page.fill("textarea[name='bio']", "This is my new bio.")

    page.click("button:has-text('Save changes')")

    # Verify redirect and message
    expect(page).to_have_url(f"{live_server.url}/profile/")
    expect(page.locator("body")).to_contain_text("Profile updated successfully.")

    # Verify values persisted
    expect(page.locator("input[name='first_name']")).to_have_value("UpdatedFirst")
    expect(page.get_by_text("This is my new bio.")).to_be_visible()


@pytest.mark.django_db
def test_user_settings_ui(page: Page, live_server, user):
    """
    Test changing user settings (language and theme).
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Go to settings
    page.goto(f"{live_server.url}/settings/")

    # Change theme to dark
    page.select_option("select[name='theme']", "dark")
    # Change language to Nederlands
    page.select_option("select[name='language']", "nl")

    page.click("button:has-text('Save changes')")

    # Verify redirection and success message
    expect(page).to_have_url(re.compile(r".*/settings/$"))

    user.refresh_from_db()
    assert user.theme == "dark"
    assert user.language == "nl"


@pytest.mark.django_db
def test_order_tracking_ui(page: Page, live_server, user, recipe):
    """
    Test that orders appear in the profile and filtering works.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")
    from decimal import Decimal

    # Recipe must have a price
    recipe.price = Decimal("15.00")
    recipe.save()

    # Create an order
    o = Order.objects.create(user=user, status="SHIPPED", total_price=Decimal("15.00"))
    OrderItem.objects.create(order=o, recipe=recipe)

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Go to profile
    page.goto(f"{live_server.url}/profile/")

    # Check that order is visible
    expect(page.locator("span.chip", has_text="Shipped")).to_be_visible()
    expect(page.get_by_text("15.00 €")).to_be_visible()
    expect(page.get_by_text(recipe.title)).to_be_visible()

    # Filter by Status
    page.select_option("select[name='status']", "PENDING")
    # Page should reload
    expect(page.get_by_text("No previous orders found.")).to_be_visible()

    # Filter back to all or shipped
    page.select_option("select[name='status']", "SHIPPED")
    expect(page.locator("span.chip", has_text="Shipped")).to_be_visible()


@pytest.mark.django_db
def test_checkout_multiple_items_ui(page: Page, live_server, user):
    """
    Test checking out with multiple items in the cart.
    """
    User.objects.create_superuser("admin", "admin@example.com", "password")
    r1 = Recipe.objects.create(  # noqa: F841
        title="Sandwich 1", price=10.00, slug="s1", is_approved=True
    )
    r2 = Recipe.objects.create(  # noqa: F841
        title="Sandwich 2", price=20.00, slug="s2", is_approved=True
    )

    # Login
    page.goto(f"{live_server.url}/login/")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password")
    page.press("input[name='password']", "Enter")

    # Add r1
    page.goto(f"{live_server.url}/recipes/s1/")
    page.click("button:has-text('Add to Cart')")

    # Add r2
    page.goto(f"{live_server.url}/recipes/s2/")
    page.click("button:has-text('Add to Cart')")

    # Go to cart
    page.goto(f"{live_server.url}/cart/")

    # Verify items are there
    expect(page.get_by_text("Sandwich 1")).to_be_visible()
    expect(page.get_by_text("Sandwich 2")).to_be_visible()
    # Flexible check for total
    expect(page.locator("body")).to_contain_text("30")
    expect(page.locator("body")).to_contain_text("€")

    # Click Checkout
    # Use a more robust locator if possible
    checkout_btn = page.get_by_role("button", name="Checkout")
    expect(checkout_btn).to_be_enabled()
    checkout_btn.click()

    # Should redirect to profile
    expect(page).to_have_url(f"{live_server.url}/profile/")
    expect(page.get_by_text("Orders submitted successfully!")).to_be_visible()

    # Verify order items
    expect(page.get_by_text("Sandwich 1")).to_be_visible()
    expect(page.get_by_text("(+1)")).to_be_visible()  # Sandwich 1 (+1 other item)
