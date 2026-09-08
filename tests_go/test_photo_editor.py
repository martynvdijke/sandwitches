"""Browser photo editor UI tests (browser-photo-editing change)."""

import sqlite3

import pytest
from playwright.sync_api import Page, expect

from conftest import create_admin, login_session
from test_photo_zoom import _png_bytes

GO_URL = "http://127.0.0.1:6279"
NAV = {"wait_until": "commit", "timeout": 15000}

DEVICES = {
    "mobile": {"width": 375, "height": 667},
    "desktop": {"width": 1280, "height": 720},
}
DEVICE_NAMES = list(DEVICES.keys())


def _recipe_id(go_server, slug) -> int:
    conn = sqlite3.connect(go_server.db_file)
    row = conn.execute("SELECT id FROM recipes WHERE slug = ?", (slug,)).fetchone()
    conn.close()
    return row[0]


def _fill_new_recipe(page: Page, go_server, title):
    page.goto(f"{go_server.url}/dashboard/recipes/add")
    page.wait_for_selector("h4", timeout=10000)
    page.fill("input[name='title']", title)
    page.fill("textarea[name='description']", "Edited photo recipe")
    page.fill("textarea[name='ingredients']", "Bread")
    page.fill("textarea[name='instructions']", "Toast it")
    page.fill("input[name='servings']", "2")
    checkbox = page.locator("input[name='is_approved']")
    if checkbox.count():
        checkbox.check()


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_new_file_edit_apply_creates_recipe(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    _fill_new_recipe(page, go_server, f"Edited {device_name}")

    page.set_input_files(
        "input[name='image']",
        {"name": "sandwich.png", "mimeType": "image/png", "buffer": _png_bytes()},
    )
    edit_btn = page.locator('[data-edit-photo="new"]')
    expect(edit_btn).to_be_visible()

    edit_btn.click()
    overlay = page.locator("#photo-editor")
    expect(overlay).to_be_visible()
    expect(page.locator("#pe-canvas")).to_be_visible()

    # Rotate + aspect switch, then apply back into the file input.
    page.locator('[data-rotate="cw"]').click()
    page.locator('[data-aspect="1"]').click()
    page.locator("[data-editor-apply]").click()
    expect(overlay).to_be_hidden()

    page.locator("button[type='submit']").click()
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=10000)
    expect(page.locator("tbody")).to_contain_text(f"Edited {device_name}")


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_existing_photo_edit_and_cancel(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    _fill_new_recipe(page, go_server, f"Existing {device_name}")
    page.set_input_files(
        "input[name='image']",
        {"name": "sandwich.png", "mimeType": "image/png", "buffer": _png_bytes()},
    )
    page.locator("button[type='submit']").click()
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=10000)

    slug = page.locator("tbody td a").first.get_attribute("href").rstrip("/").rsplit("/", 1)[-1]
    rid = _recipe_id(go_server, slug)
    page.goto(f"{go_server.url}/dashboard/recipes/{rid}/edit", **NAV)

    # Existing-photo Edit button is revealed by the editor script.
    edit_btn = page.locator('[data-edit-photo="existing"]')
    expect(edit_btn).to_be_visible()

    # Cancel discards: overlay closes, form still submits fine.
    edit_btn.click()
    overlay = page.locator("#photo-editor")
    expect(overlay).to_be_visible()
    page.locator("[data-editor-close]").click()
    expect(overlay).to_be_hidden()

    # Re-open, flip + reset, apply, save.
    edit_btn.click()
    page.locator('[data-flip="h"]').click()
    page.locator("[data-reset]").click()
    page.locator("[data-editor-apply]").click()
    expect(overlay).to_be_hidden()
    page.locator("button[type='submit']").click()
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=10000)


def test_nojs_fallback_hides_editor(page: Page, go_server, browser):
    create_admin(go_server)
    ctx = browser.new_context(java_script_enabled=False)
    nojs = ctx.new_page()
    try:
        login_session(nojs, go_server, "admin", "adminpass123")
        _fill_new_recipe(nojs, go_server, "NoJS Sandwich")
        # No Edit affordance without JS; plain upload path untouched.
        expect(nojs.locator("[data-edit-photo]:visible")).to_have_count(0)
        expect(nojs.locator("#photo-editor")).to_be_hidden()
        nojs.set_input_files(
            "input[name='image']",
            {"name": "sandwich.png", "mimeType": "image/png", "buffer": _png_bytes()},
        )
        nojs.locator("button[type='submit']").click()
        nojs.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=10000)
        expect(nojs.locator("tbody")).to_contain_text("NoJS Sandwich")
    finally:
        ctx.close()
