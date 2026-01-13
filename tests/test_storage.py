from sandwitches.storage import is_database_readable
from unittest.mock import patch
from pathlib import Path  # Import Path


def test_is_database_file_readable_when_exists(tmp_path):
    f = tmp_path / "db.sqlite3"
    f.write_text("x")
    assert is_database_readable(f) is True


def test_is_database_file_readable_nonexistent(tmp_path):
    f = tmp_path / "nope.sqlite3"
    assert is_database_readable(f) is False


def test_is_database_file_not_readable_permissions(tmp_path):
    f = tmp_path / "db.sqlite3"
    f.write_text("x")
    # No need to remove permissions, as we are mocking builtins.open

    # Define a custom side_effect for builtins.open
    original_open = open  # Keep a reference to the original builtins.open

    def custom_mock_open(file, mode="r", *args, **kwargs):
        if (
            Path(file) == f and "r" in mode
        ):  # If it's our test file and we're trying to read
            raise PermissionError(
                "Simulated permission denied"
            )  # Simulate permission denied
        return original_open(
            file, mode, *args, **kwargs
        )  # For all other files, use original open

    with patch("builtins.open", side_effect=custom_mock_open):
        assert is_database_readable(f) is False

    # The original test restored permissions, which is good practice
    # even if not strictly needed in this specific mocked test.
    # The file still exists on the filesystem and might have had permissions changed by previous (failed) test runs.
    f.chmod(0o644)


def test_default_uses_settings(monkeypatch, tmp_path):
    f = tmp_path / "db.sqlite3"
    f.write_text("x")
    import django.conf

    # set the settings.DATABASE_FILE to our temp file and ensure the function picks it up
    monkeypatch.setattr(django.conf.settings, "DATABASE_FILE", f)
    assert is_database_readable(None) is True
