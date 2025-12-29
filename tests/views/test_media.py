import pytest
from django.urls import reverse


@pytest.fixture
def media_root(tmp_path, settings):
    settings.MEDIA_ROOT = tmp_path
    return tmp_path


@pytest.mark.django_db
def test_media_valid_file(client, media_root):
    # Create a dummy image file in media root
    sub = media_root / "uploads"
    sub.mkdir()
    f = sub / "test.jpg"
    f.write_bytes(b"fakeimagecontent")

    url = reverse("media", kwargs={"file_path": "uploads/test.jpg"})
    response = client.get(url)

    # Check if we get 200 and content
    assert response.status_code == 200
    assert b"fakeimagecontent" in response.getvalue()


@pytest.mark.django_db
def test_media_file_not_found(client, media_root):
    url = reverse("media", kwargs={"file_path": "non_existent.jpg"})
    response = client.get(url)
    assert response.status_code == 404


@pytest.mark.django_db
def test_media_directory_traversal(client, media_root):
    # Try to access a file outside media root
    # Create a secret file outside
    secret = media_root.parent / "secret.txt"
    secret.write_bytes(b"secret")

    # Attempt to access via ../
    # Django url pattern might catch this, or the view logic
    # path:file_path usually matches everything including slashes.
    # But usually webservers/clients normalize paths.
    # We pass it as string to reverse, but client.get might normalize.
    # We'll try to construct URL carefully.

    # If we use reverse, it might urlencode ".."
    url = reverse("media", kwargs={"file_path": "../secret.txt"})

    # Note: the view logic:
    # base_path.joinpath(file_path).resolve()
    # if base_path not in full_path.parents: return BadRequest

    response = client.get(url)

    # Should be 400 Bad Request (Access Denied) or 404 depending on exact resolution
    assert response.status_code in [400, 404]


@pytest.mark.django_db
def test_media_invalid_content_type(client, media_root):
    # Create a text file
    f = media_root / "test.txt"
    f.write_bytes(b"text content")

    url = reverse("media", kwargs={"file_path": "test.txt"})
    response = client.get(url)

    # View checks if content_type starts with "image/"
    # mimetypes.guess_type("test.txt") -> "text/plain"
    assert response.status_code == 400
