import hashlib
import os
from django.core.files.base import ContentFile
from django.core.files.storage import FileSystemStorage


class HashedFilenameStorage(FileSystemStorage):
    """
    Save uploaded files under a hash of their contents + original extension.
    Example output: media/recipes/3f8a9d...png
    """

    def _save(self, name, content):
        # ensure we read bytes
        try:
            content.seek(0)
        except Exception:
            pass

        data = content.read()
        # compute hash (use first 32 hex chars to keep names shorter)
        h = hashlib.sha256(data).hexdigest()[:32]
        ext = os.path.splitext(name)[1].lower() or ""
        # store all recipe images under recipes/
        name = f"recipes/{h}{ext}"

        # wrap bytes into a ContentFile so Django storage works consistently
        content = ContentFile(data)
        return super()._save(name, content)
