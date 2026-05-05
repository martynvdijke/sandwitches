import re

import pytest
from playwright.sync_api import Page, expect

from conftest import (
    create_admin,
    create_recipe_via_ui,
    login_session,
)

GO_URL = "http://127.0.0.1:6279"
# Use domcontentloaded to avoid waiting for large static assets (2MB main.js)
NAV_OPTS = {"wait_until": "commit", "timeout": 15000}  # avoid waiting for large static assets


def test_initial_setup_redirect(page: Page):
    """Fresh DB: visiting / should redirect to /setup."""
    page.goto(GO_URL)
    expect(page).to_have_url(f"{GO_URL}/setup")
    heading = page.get_by_role("heading", name="Admin Setup")
    expect(heading).to_be_visible()


def test_setup_creates_admin(page: Page, go_server):
    """Complete the setup flow and verify admin is created."""
    page.goto(f"{GO_URL}/setup")
    page.fill("input[name='username']", "admin")
    page.fill("input[name='email']", "admin@test.com")
    page.fill("input[name='password1']", "adminpass123")
    page.fill("input[name='password2']", "adminpass123")
    page.get_by_role("button", name="Create Admin").click()

    # After setup, redirect to /dashboard
    expect(page).to_have_url(re.compile(r".*/dashboard$"))

    # Visiting setup again redirects to / (admin already exists)
    page.goto(f"{GO_URL}/setup")
    expect(page).to_have_url(f"{GO_URL}/")


def test_login_flow(page: Page, go_server):
    """Test the login flow and redirection to the index page."""
    create_admin(go_server)

    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='email']", "test@example.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    # Logout and login again
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/login")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='password']", "password123")
    page.locator("input[name='password']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    expect(page.locator("a[href='/profile']")).to_be_visible()


def test_recipe_favoriting(page: Page, go_server):
    """Test favoriting a recipe."""
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "Test Sandwich", "Delicious test sandwich",
                         "Bread\nCheese", "Put cheese on bread.", servings=1, price="10.00")

    # Get recipe slug from admin recipe list
    page.goto(f"{GO_URL}/dashboard/recipes")
    recipe_link = page.locator("tbody td a").first
    recipe_href = recipe_link.get_attribute("href")
    slug = recipe_href.rstrip("/").rsplit("/", 1)[-1]

    # Signup as regular user then favorite
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "testuser")
    page.fill("input[name='email']", "test@example.com")
    page.fill("input[name='first_name']", "Test")
    page.fill("input[name='last_name']", "User")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV_OPTS)

    favorite_link = page.locator("a[href*='/recipes/favorite/']")
    expect(favorite_link).to_be_visible()
    favorite_link.click()
    page.wait_for_url(re.compile(r".*/recipes/"), timeout=5000)


def test_recipe_rating(page: Page, go_server):
    """Test rating and commenting on a recipe."""
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "RateMe Sandwich", "A recipe for rating",
                         "1 cup Flour", "Mix and bake", servings=2)

    # Get slug
    page.goto(f"{GO_URL}/dashboard/recipes")
    recipe_link = page.locator("tbody td a").first
    slug = (recipe_link.get_attribute("href") or "").rstrip("/").rsplit("/", 1)[-1]

    # Logout and create regular user
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "rater")
    page.fill("input[name='email']", "rater@test.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    # Rate
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV_OPTS)
    comment_text = "This is a fantastic test sandwich!"
    page.fill("textarea[name='comment']", comment_text)
    page.locator("input[name='score']").fill("8")
    page.click("button:has-text('Submit Rating')")
    page.wait_for_url(re.compile(r".*/recipes/" + re.escape(slug)), timeout=5000)
    expect(page.get_by_text(comment_text).first).to_be_visible()


def test_signup_flow(page: Page, go_server):
    """Test the user signup process."""
    create_admin(go_server)

    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='first_name']", "New")
    page.fill("input[name='last_name']", "User")
    page.fill("input[name='username']", "new_user_123")
    page.fill("input[name='email']", "new@example.com")
    page.fill("input[name='password1']", "SecurePass123!")
    page.fill("input[name='password2']", "SecurePass123!")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")
    expect(page.locator("a[href='/profile']")).to_be_visible()


def test_order_sandwich_ui(page: Page, go_server):
    """Test ordering a sandwich via Add to Cart."""
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "OrderMe Sandwich", "Yummy",
                         "Bread", "Toast bread", servings=1, price="10.00")

    # Get slug
    page.goto(f"{GO_URL}/dashboard/recipes")
    recipe_link = page.locator("tbody td a").first
    slug = (recipe_link.get_attribute("href") or "").rstrip("/").rsplit("/", 1)[-1]

    # Signup as customer
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "customer1")
    page.fill("input[name='email']", "cust@test.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    # Add to cart
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV_OPTS)
    add_to_cart = page.locator("a:has-text('Add to Cart')")
    expect(add_to_cart).to_be_visible()
    add_to_cart.click(no_wait_after=True)
    page.wait_for_url(f"{GO_URL}/cart", timeout=5000)
    expect(page.get_by_text("OrderMe Sandwich")).to_be_visible()


def test_update_profile_ui(page: Page, go_server):
    """Test updating user profile bio and names."""
    create_admin(go_server)
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "profileuser")
    page.fill("input[name='email']", "prof@test.com")
    page.fill("input[name='first_name']", "OriginalFirst")
    page.fill("input[name='last_name']", "OriginalLast")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    page.goto(f"{GO_URL}/profile")
    expect(page.locator("input[name='first_name']")).to_have_value("OriginalFirst")

    page.fill("input[name='first_name']", "UpdatedFirst")
    page.fill("input[name='last_name']", "UpdatedLast")
    page.fill("textarea[name='bio']", "This is my new bio.")
    page.click("button:has-text('Save changes')")
    expect(page).to_have_url(f"{GO_URL}/profile")
    expect(page.locator("input[name='first_name']")).to_have_value("UpdatedFirst")
    expect(page.locator("textarea[name='bio']")).to_have_value("This is my new bio.")


def test_user_settings_ui(page: Page, go_server):
    """Test changing user settings (language and theme)."""
    create_admin(go_server)
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "settingsuser")
    page.fill("input[name='email']", "set@test.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    page.goto(f"{GO_URL}/settings")
    page.select_option("select[name='theme']", "dark")
    page.select_option("select[name='language']", "nl")
    page.click("button:has-text('Save Settings')")
    expect(page).to_have_url(f"{GO_URL}/settings", timeout=5000)
    expect(page.locator("select[name='theme']")).to_have_value("dark")
    expect(page.locator("select[name='language']")).to_have_value("nl")


def test_cart_checkout_ui(page: Page, go_server):
    """Test adding items to cart and checking out."""
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "MultiItem Sandwich", "Test", "Stuff",
                         "Make it", servings=1, price="30.00")

    # Get slug
    page.goto(f"{GO_URL}/dashboard/recipes")
    slug = (page.locator("tbody td a").first.get_attribute("href") or "").rstrip("/").rsplit("/", 1)[-1]

    # Signup as customer
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "checkoutuser")
    page.fill("input[name='email']", "check@test.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    # Add to cart and checkout
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV_OPTS)
    page.locator("a:has-text('Add to Cart')").click(no_wait_after=True)
    page.wait_for_url(f"{GO_URL}/cart", timeout=5000)

    expect(page.get_by_text("MultiItem Sandwich")).to_be_visible()
    page.locator("button:has-text('Checkout')").click(timeout=5000)
    page.wait_for_url(f"{GO_URL}/profile", timeout=10000)
    # Flash message auto-hides after 5s, check heading instead
    expect(page.get_by_role("heading", name="Edit Profile")).to_be_visible()


def test_order_tracking_ui(page: Page, go_server):
    """Test order appears on profile and order detail page."""
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "Trackable Sandwich", "For tracking",
                         "Bread", "Toast", servings=1, price="15.00")

    # Get slug
    page.goto(f"{GO_URL}/dashboard/recipes")
    slug = (page.locator("tbody td a").first.get_attribute("href") or "").rstrip("/").rsplit("/", 1)[-1]

    # Signup and order
    page.goto(f"{GO_URL}/logout")
    page.goto(f"{GO_URL}/signup")
    page.fill("input[name='username']", "trackuser")
    page.fill("input[name='email']", "track@test.com")
    page.fill("input[name='password1']", "password123")
    page.fill("input[name='password2']", "password123")
    page.locator("input[name='password2']").press("Enter")
    expect(page).to_have_url(f"{GO_URL}/")

    page.goto(f"{GO_URL}/recipes/{slug}", **NAV_OPTS)
    page.locator("a:has-text('Add to Cart')").click(no_wait_after=True)
    page.wait_for_url(f"{GO_URL}/cart", timeout=5000)

    page.locator("button:has-text('Checkout')").click(timeout=5000)
    page.wait_for_url(f"{GO_URL}/profile", timeout=10000)

    expect(page.get_by_text("PENDING")).to_be_visible()
    expect(page.locator("a:has-text('View')")).to_be_visible()
