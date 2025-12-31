import pytest
from django.contrib.auth import get_user_model
from django.core.files.uploadedfile import SimpleUploadedFile

User = get_user_model()


@pytest.mark.django_db
def test_user_creation_with_extra_fields(user_factory):
    # Create user with avatar and bio
    small_gif = (
        b"\x47\x49\x46\x38\x39\x61\x01\x00\x01\x00\x80\x00\x00\x00\x00\x00"
        b"\xff\xff\xff\x21\xf9\x04\x01\x00\x00\x00\x00\x2c\x00\x00\x00\x00"
        b"\x01\x00\x01\x00\x00\x02\x01\x44\x00\x3b"
    )
    avatar = SimpleUploadedFile("avatar.gif", small_gif, content_type="image/gif")

    # create_user on Custom User model should accept extra fields if they are in the model,
    # but AbstractUser.objects.create_user typically only takes username, email, password and extra fields as kwargs.
    user = User.objects.create_user(
        username="newuser", password="password", bio="Hello", avatar=avatar
    )

    assert user.bio == "Hello"
    assert user.avatar
    # imagekit might need the file to be saved/processed
    assert user.avatar_thumbnail
    assert user.avatar_thumbnail.url


@pytest.mark.django_db
def test_user_favorites(user_factory):
    from sandwitches.models import Recipe

    user = user_factory()
    recipe = Recipe.objects.create(title="Toast", description="Bread", uploaded_by=user)

    user.favorites.add(recipe)

    assert recipe in user.favorites.all()
    assert user in recipe.favorited_by.all()
