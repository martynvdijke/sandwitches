from sandwitches.storage import is_database_readable


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
    # remove all permissions
    f.chmod(0)
    try:
        assert is_database_readable(f) is False
    finally:
        # restore permissions so tmp cleanup works
        f.chmod(0o644)


def test_default_uses_settings(monkeypatch, tmp_path):
    f = tmp_path / "db.sqlite3"
    f.write_text("x")
    import django.conf

    # set the settings.DATABASE_FILE to our temp file and ensure the function picks it up
    monkeypatch.setattr(django.conf.settings, "DATABASE_FILE", f)
    assert is_database_readable(None) is True
