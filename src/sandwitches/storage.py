import hashlib
import os
from django.core.files.base import ContentFile
from django.core.files.storage import FileSystemStorage
from pathlib import Path
from django.conf import settings
import logging


class HashedFilenameStorage(FileSystemStorage):
    """
    Save uploaded files under a hash of their contents + original extension.
    Example output: media/recipes/3f8a9d...png
    """

    def _save(self, name, content):
        try:
            content.seek(0)
        except Exception:
            pass

        data = content.read()
        h = hashlib.sha256(data).hexdigest()[:32]
        ext = os.path.splitext(name)[1].lower() or ""
        name = f"recipes/{h}{ext}"

        content = ContentFile(data)
        return super()._save(name, content)


def is_database_readable(path=None) -> bool:
    """
    Check whether the database file exists and is readable.

    If `path` is None, uses `django.conf.settings.DATABASE_FILE` by default.

    Returns:
        bool: True if the file exists and is readable by the current process, False otherwise.
    """

    if path is None:
        path = getattr(settings, "DATABASE_FILE", None)

    if not path:
        return False

    p = Path(path)
    logging.debug(f"Checking database file readability at: {p}")
    try:
        if p.is_file():
            with open(p, "r"):  # Removed 'as f'
                # If we can open it, it's readable. No need to read content.
                pass
            logging.debug(f"Database file at {p} is readable.")
            return True
        else:
            logging.error(f"Database file at {p} does not exist.")
            return False
    except IOError:
        logging.error(
            f"Database file at {p} is not readable due to permission or other IO error."
        )
        return False


def is_database_writable(path=None) -> bool:
    """
    Check whether the database file exists and is writable.

    If `path` is None, uses `django.conf.settings.DATABASE_FILE` by default.

    Returns:
        bool: True if the file exists and is writable by the current process, False otherwise.
    """

    if path is None:
        path = getattr(settings, "DATABASE_FILE", None)

    if not path:
        return False

    p = Path(path)
    logging.debug(f"Checking database file writability at: {p}")
    try:
        # If the file exists, try to open it for appending
        if p.is_file():
            with open(p, "a"):  # Removed 'as f'
                pass
            logging.debug(f"Database file at {p} is writable.")
            return True
        else:
            # If file does not exist, check if its parent directory is writable
            if p.parent.is_dir() and os.access(p.parent, os.W_OK):
                # Try creating a dummy file to confirm writability
                dummy_file = p.parent / f".tmp_writable_test_{os.getpid()}"
                try:
                    dummy_file.touch()
                    dummy_file.unlink()
                    logging.debug(f"Database path at {p.parent} is writable.")
                    return True
                except IOError:
                    logging.error(f"Cannot create dummy file in {p.parent}.")
                    return False
            else:
                logging.error(
                    f"Parent directory {p.parent} is not writable or does not exist."
                )
                return False

    except IOError:
        logging.error(
            f"Database file at {p} is not writable due to permission or other IO error."
        )
        return False
