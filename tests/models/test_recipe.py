import pytest
from unittest.mock import patch
from sandwitches.models import Recipe, Tag


@pytest.mark.django_db
def test_recipe_slug_generation():
    r1 = Recipe.objects.create(title="My Sandwich")
    assert r1.slug == "my-sandwich"

    # Use a title that is different but slugifies to the same string
    r2 = Recipe.objects.create(title="My Sandwich!")
    assert r2.slug == "my-sandwich-1"


@pytest.mark.django_db
def test_set_tags_from_string():
    r = Recipe.objects.create(title="Tag Test")

    # Create one existing tag
    Tag.objects.create(name="Spicy")

    r.set_tags_from_string("Spicy,  Sweet , savory ")

    tags = r.tags.all().order_by("name")
    tag_names = [t.name for t in tags]

    assert "Spicy" in tag_names
    assert "Sweet" in tag_names
    assert "savory" in tag_names
    assert len(tag_names) == 3

    # Verify new tags were created
    assert Tag.objects.count() == 3


@pytest.mark.django_db
def test_recipe_tag_list():
    r = Recipe.objects.create(title="List Test")
    r.set_tags_from_string("A, B")

    tags = r.tag_list()
    assert "A" in tags
    assert "B" in tags
    assert len(tags) == 2


@pytest.mark.django_db
def test_recipe_str():
    r = Recipe.objects.create(title="Tasty Sub")
    assert str(r) == "Tasty Sub"


@pytest.mark.django_db
def test_recipe_get_absolute_url():
    r = Recipe.objects.create(title="Url Test")
    # Assuming url pattern is /recipe/<slug>/ based on reverse call in model
    # We might not have the full urlconf loaded in unit tests easily without checking urls.py
    # but get_absolute_url should return something.
    url = r.get_absolute_url()
    assert r.slug in url


@pytest.mark.django_db
def test_recipe_creation_triggers_email_task(settings):
    # Mock the email_users task object
    settings.SEND_EMAIL = True

    with patch("sandwitches.models.email_users") as mock_task:
        r = Recipe.objects.create(title="Email Test Recipe")
        mock_task.enqueue.assert_called_once_with(recipe_id=r.pk)


@pytest.mark.django_db
def test_recipe_update_does_not_trigger_email_task(settings):
    settings.SEND_EMAIL = True

    with patch("sandwitches.models.email_users") as mock_task:
        r = Recipe.objects.create(title="Update Test Recipe")
        mock_task.enqueue.assert_called_once()  # Called on create

        mock_task.enqueue.reset_mock()

        r.description = "Updated description"
        r.save()

        mock_task.enqueue.assert_not_called()


@pytest.mark.django_db
def test_recipe_creation_no_email_if_disabled(settings):
    settings.SEND_EMAIL = False

    with patch("sandwitches.models.email_users") as mock_task:
        Recipe.objects.create(title="No Email Test")
        mock_task.enqueue.assert_not_called()
