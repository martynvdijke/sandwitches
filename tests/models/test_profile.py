import pytest
from sandwitches.models import Profile

from django.core.files.uploadedfile import SimpleUploadedFile


@pytest.mark.django_db
def test_profile_creation(user_factory):
    user = user_factory()

    # Create profile with avatar
    # Use valid minimal GIF data
    small_gif = (
        b"\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00"
        b"\xff\xff\xff\x21\xf9\x04\x01\x00\x00\x00\x00\x2c\x00\x00\x00\x00"
        b"\x01\x00\x01\x00\x00\x02\x01\x44\x00\x3b"
    )
    avatar = SimpleUploadedFile("avatar.gif", small_gif, content_type="image/gif")
    profile = Profile.objects.create(user=user, bio="Hello", avatar=avatar)

    assert profile.bio == "Hello"
    assert str(profile) == f"{user.username}'s Profile"

    # Now avatar_thumbnail should be truthy (or at least source file exists)
    assert profile.avatar_thumbnail
    assert profile.avatar_thumbnail.url
