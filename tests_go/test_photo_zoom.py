"""Photo zoom lightbox UI tests (add-photo-zoom change)."""

import re
import struct
import zlib

import pytest
from playwright.sync_api import Page, expect

from conftest import (
    create_admin,
    create_recipe_via_ui,
    get_slug_from_admin_list,
    login_session,
)

GO_URL = "http://127.0.0.1:6279"
NAV = {"wait_until": "commit", "timeout": 15000}

DEVICES = {
    "mobile": {"width": 375, "height": 667},
    "desktop": {"width": 1280, "height": 720},
}
DEVICE_NAMES = list(DEVICES.keys())


def _png_bytes(width=600, height=400) -> bytes:
    """Solid-color PNG, generated with the stdlib only."""

    def chunk(tag: bytes, data: bytes) -> bytes:
        return (
            struct.pack(">I", len(data))
            + tag
            + data
            + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)
        )

    ihdr = struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0)
    row = b"\x00" + b"\x30\x90\xc0" * width
    idat = zlib.compress(row * height)
    return (
        b"\x89PNG\r\n\x1a\n"
        + chunk(b"IHDR", ihdr)
        + chunk(b"IDAT", idat)
        + chunk(b"IEND", b"")
    )


def _create_recipe_with_image(page: Page, go_server, title="ZoomMe Sandwich"):
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    page.goto(f"{go_server.url}/dashboard/recipes/add")
    page.wait_for_selector("h4", timeout=10000)
    page.fill("input[name='title']", title)
    page.fill("textarea[name='description']", "A recipe with a photo")
    page.fill("textarea[name='ingredients']", "Bread")
    page.fill("textarea[name='instructions']", "Toast it")
    page.fill("input[name='servings']", "2")
    checkbox = page.locator("input[name='is_approved']")
    if checkbox.count():
        checkbox.check()
    page.set_input_files(
        "input[name='image']",
        {"name": "sandwich.png", "mimeType": "image/png", "buffer": _png_bytes()},
    )
    page.locator("button[type='submit']").click()
    page.wait_for_url(f"{go_server.url}/dashboard/recipes", timeout=5000)
    slug = get_slug_from_admin_list(page, go_server)
    # Admin-created recipes get no uploaded_by_id (pre-existing app behavior),
    # which excludes them from the home grid's admin-group JOIN. Patch it so
    # the recipe card (and its zoom button) renders on the home page.
    import sqlite3
    conn = sqlite3.connect(go_server.db_file)
    conn.execute(
        "UPDATE recipes SET uploaded_by_id = (SELECT id FROM users WHERE username = 'admin') "
        "WHERE slug = ?",
        (slug,),
    )
    conn.commit()
    conn.close()
    return slug


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_hero_click_opens_lightbox(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    slug = _create_recipe_with_image(page, go_server)
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV)

    hero = page.locator("#photo-zoom-trigger, img[data-zoom-src]").first
    expect(hero).to_be_visible()
    dialog = page.locator("#photo-zoom")
    expect(dialog).not_to_have_attribute("open", re.compile(r".*"))

    hero.click()
    expect(dialog).to_have_attribute("open", re.compile(r".*"))
    lightbox_img = page.locator("#photo-zoom-img")
    expect(lightbox_img).to_be_visible()
    src = lightbox_img.get_attribute("src") or ""
    assert "/thumb/" in src and "w=1600" in src


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_escape_and_backdrop_close_lightbox(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    slug = _create_recipe_with_image(page, go_server)
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV)

    dialog = page.locator("#photo-zoom")
    page.locator("img[data-zoom-src]").first.click()
    expect(dialog).to_have_attribute("open", re.compile(r".*"))

    # Escape closes
    page.keyboard.press("Escape")
    expect(dialog).not_to_have_attribute("open", re.compile(r".*"))

    # Re-open, then backdrop click closes
    page.locator("img[data-zoom-src]").first.click()
    expect(dialog).to_have_attribute("open", re.compile(r".*"))
    dialog.click(position={"x": 5, "y": 5})
    expect(dialog).not_to_have_attribute("open", re.compile(r".*"))


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_card_zoom_button_does_not_navigate(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    slug = _create_recipe_with_image(page, go_server)
    page.goto(f"{GO_URL}/", **NAV)

    card_btn = page.locator(".zoom-btn[data-zoom-src]").first
    expect(card_btn).to_be_visible()
    card_btn.click()

    # Lightbox opens, no navigation away from home
    expect(page).to_have_url(f"{GO_URL}/")
    dialog = page.locator("#photo-zoom")
    expect(dialog).to_have_attribute("open", re.compile(r".*"))
    expect(page.locator("#photo-zoom-img")).to_be_visible()


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_wheel_zoom_and_pan(page: Page, go_server, device_name):
    if device_name == "mobile":
        pytest.skip("wheel/pan interaction asserted on desktop only")
    page.set_viewport_size(DEVICES["desktop"])
    slug = _create_recipe_with_image(page, go_server)
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV)

    page.locator("img[data-zoom-src]").first.click()
    img = page.locator("#photo-zoom-img")
    expect(img).to_be_visible()

    before = img.evaluate("el => el.style.transform || ''")
    box = page.locator("#photo-zoom-stage").bounding_box()
    cx, cy = box["x"] + box["width"] / 2, box["y"] + box["height"] / 2
    page.mouse.move(cx, cy)
    page.mouse.wheel(0, -240)  # zoom in
    after = img.evaluate("el => el.style.transform || ''")
    assert after != before
    assert "scale(" in after

    # Pan while zoomed changes translation
    tx_before = img.evaluate("el => el.style.transform")
    page.mouse.down()
    page.mouse.move(cx + 60, cy + 40, steps=5)
    page.mouse.up()
    tx_after = img.evaluate("el => el.style.transform")
    assert tx_after != tx_before


@pytest.mark.parametrize("device_name", DEVICE_NAMES)
def test_placeholder_not_zoomable(page: Page, go_server, device_name):
    page.set_viewport_size(DEVICES[device_name])
    create_admin(go_server)
    login_session(page, go_server, "admin", "adminpass123")
    create_recipe_via_ui(page, go_server, "NoPhoto Sandwich", "No image",
                         "Bread", "Toast it", servings=2)
    slug = get_slug_from_admin_list(page, go_server)
    page.goto(f"{GO_URL}/recipes/{slug}", **NAV)

    # No zoom triggers anywhere on the detail page
    expect(page.locator("[data-zoom-src]")).to_have_count(0)
    dialog = page.locator("#photo-zoom")
    expect(dialog).not_to_have_attribute("open", re.compile(r".*"))
