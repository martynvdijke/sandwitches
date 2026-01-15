import pytest
from playwright.sync_api import Page, expect


@pytest.mark.django_db
def test_homepage(page: Page, live_server):
    page.goto(live_server.url)
    expect(page).to_have_title("Initial setup — Create admin")

    heading = page.get_by_role("heading", name="Create initial administrator")
    expect(heading).to_be_visible()
