import pytest
import re
from playwright.sync_api import Page, expect
from sandwitches.models import User, Recipe

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
        instructions="Put cheese on bread."
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

    # Should redirect to index (fixed in settings)
    expect(page).to_have_url(f"{live_server.url}/")
    
    # Verify user is logged in by checking for the user menu avatar
    expect(page.locator("img[data-ui='#user-menu']")).to_be_visible()

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
    expect(toggle_btn).to_be_visible() # Ensure we are logged in
    expect(toggle_btn.locator("i")).to_have_text("favorite_border")
    
    # Click favorite
    toggle_btn.click()
    
    # Verify state change (icon is favorite)
    # The page reloads or updates. We expect the icon text to change.
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
    # 'Comment' is the label for the textarea, but let's use name to be safe
    page.fill("textarea[name='comment']", comment_text)
    
    page.click("button:has-text('Submit Rating')")
    
    # Verify the rating appears on the page
    # Use .first to avoid strict mode violation if text appears in textarea and display
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
    # Use a stronger password to avoid validation errors
    strong_pass = "SecurePass123!" 
    page.fill("input[name='password1']", strong_pass)
    page.fill("input[name='password2']", strong_pass)
    
    # Submit via Enter on last field
    page.press("input[name='password2']", "Enter")
    
    # Should redirect to index and be logged in
    expect(page).to_have_url(f"{live_server.url}/")
    expect(page.locator("img[data-ui='#user-menu']")).to_be_visible()