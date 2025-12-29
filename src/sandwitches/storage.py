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
    readable = p.is_file() and os.access(p, os.R_OK)
    if not readable:
        logging.error(f"Database file at {p} is not readable or does not exist.")
        return False
    else:
        logging.debug(f"Database file at {p} is readable.")
        return readable


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
    writable = p.is_file() and os.access(p, os.W_OK)
    if not writable:
        logging.error(f"Database file at {p} is not writable or does not exist.")
        return False
    else:
        logging.debug(f"Database file at {p} is writable.")
    return writable
