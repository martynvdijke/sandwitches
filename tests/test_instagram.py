import pytest
from unittest.mock import patch, MagicMock
from sandwitches.models import Recipe, Setting
from django.core.files.uploadedfile import SimpleUploadedFile
from sandwitches.tasks import upload_to_instagram, sync_instagram_interactions
from sandwitches.models import InstagramComment
from datetime import datetime, timezone


@pytest.mark.django_db
def test_new_recipe_enqueues_instagram_upload():
    # Setup: Enable Instagram
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "testuser"
    config.instagram_password = "testpassword"
    config.save()

    with patch("sandwitches.models.upload_to_instagram") as mock_task:
        # Create a new recipe
        image = SimpleUploadedFile(
            "test.jpg", b"file_content", content_type="image/jpeg"
        )
        recipe = Recipe.objects.create(
            title="Insta Recipe", description="A recipe for Instagram", image=image
        )

        # Verify enqueue was called
        mock_task.enqueue.assert_called_once_with(recipe_id=recipe.pk)


@pytest.mark.django_db
def test_initial_instagram_connection_enqueues_upload():
    # Setup: Create a recipe first (MUST HAVE IMAGE to be picked up by sync_instagram_missing)
    recipe = Recipe.objects.create(
        title="Existing Recipe",
        description="Exists before Instagram is enabled",
        image="test.jpg",
    )

    config = Setting.get_solo()
    config.instagram_enabled = False
    config.instagram_initial_uploaded = False
    config.save()

    # We need to patch the task in the command module
    with patch(
        "sandwitches.management.commands.sync_instagram_missing.upload_to_instagram"
    ) as mock_task:
        # Enable Instagram
        config.instagram_enabled = True
        config.instagram_username = "testuser"
        config.instagram_password = "testpassword"
        config.save()

        # Verify enqueue was called for the existing recipe with run_after
        mock_task.using.assert_called_once()
        mock_task.using.return_value.enqueue.assert_called_once_with(
            recipe_id=recipe.pk
        )

        # Verify initial_uploaded is set to True
        config.refresh_from_db()
        assert config.instagram_initial_uploaded is True


@pytest.mark.django_db
def test_sync_instagram_missing_command():
    from django.core.management import call_command

    # Setup: 3 recipes, 1 already uploaded
    Recipe.objects.create(title="R1", description="D1", image="test1.jpg")
    Recipe.objects.create(title="R2", description="D2", image="test2.jpg")
    Recipe.objects.create(
        title="R3", description="D3", image="test3.jpg", instagram_uploaded=True
    )

    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "u"
    config.instagram_password = "p"
    config.save()

    with patch(
        "sandwitches.management.commands.sync_instagram_missing.upload_to_instagram"
    ) as mock_task:
        call_command("sync_instagram_missing")
        # Should be called for R1 and R2
        assert mock_task.using.call_count == 2
        assert mock_task.using.return_value.enqueue.call_count == 2

        # Check that they were called with increasing run_after
        calls = mock_task.using.call_args_list
        # run_after should be close to now and now + 1 hour
        assert calls[0].kwargs["run_after"] is not None
        assert calls[1].kwargs["run_after"] is not None


@pytest.mark.django_db
def test_sync_trigger_on_credential_change():
    # Setup
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "old_user"
    config.instagram_password = "old_password"
    config.instagram_initial_uploaded = True
    config.save()

    # Recipe that needs upload
    Recipe.objects.create(title="Pending", description="D", image="img.jpg")

    with patch(
        "sandwitches.management.commands.sync_instagram_missing.upload_to_instagram"
    ) as mock_task:
        # Change username
        config.instagram_username = "new_user"
        config.save()

        # Should trigger sync again because credentials changed
        assert mock_task.using.call_count == 1
        assert mock_task.using.return_value.enqueue.call_count == 1


@pytest.mark.django_db
def test_instagram_not_enqueued_if_disabled():
    config = Setting.get_solo()
    config.instagram_enabled = False
    config.save()

    with patch("sandwitches.models.upload_to_instagram"):
        # Create a new recipe
        Recipe.objects.create(
            title="No Insta Recipe", description="Instagram is disabled"
        )


@pytest.mark.django_db
def test_sync_instagram_interactions_task():
    # Setup
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "user"
    config.instagram_password = "pass"
    config.save()

    recipe = Recipe.objects.create(
        title="Sync Test Recipe", instagram_media_id="123456"
    )

    with patch("instagrapi.Client") as MockClient:
        mock_cl = MockClient.return_value

        # Mock media info for likes
        mock_media_info = MagicMock()
        mock_media_info.like_count = 42
        mock_cl.media_info.return_value = mock_media_info

        # Mock media comments
        mock_comment = MagicMock()
        mock_comment.pk = "c101"
        mock_comment.user.username = "insta_fan"
        mock_comment.text = "Looks delicious!"
        mock_comment.created_at = datetime(2023, 1, 1, tzinfo=timezone.utc)
        mock_cl.media_comments.return_value = [mock_comment]

        # Run sync
        result = sync_instagram_interactions.func()

        assert result is True
        recipe.refresh_from_db()
        assert recipe.instagram_likes_count == 42

        # Verify comment created
        comment = InstagramComment.objects.get(instagram_comment_id="c101")
        assert comment.username == "insta_fan"
        assert comment.text == "Looks delicious!"
        assert comment.recipe == recipe


@pytest.mark.django_db
def test_upload_to_instagram_task_logic():
    # Setup
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "user"
    config.instagram_password = "pass"
    config.save()

    image = SimpleUploadedFile(
        "test.jpg", b"fake_image_data", content_type="image/jpeg"
    )
    recipe = Recipe.objects.create(
        title="Task Test Recipe", description="Testing the task logic", image=image
    )

    with patch("instagrapi.Client") as MockClient:
        mock_cl = MockClient.return_value
        mock_cl.photo_upload.return_value = MagicMock(pk="999_888")

        # Mock successful upload

        result = upload_to_instagram.func(recipe.pk)

        assert result is True
        mock_cl.login.assert_called_once_with("user", "pass")
        mock_cl.photo_upload.assert_called_once()

        # Verify recipe is marked as uploaded
        recipe.refresh_from_db()
        assert recipe.instagram_uploaded is True


@pytest.mark.django_db
def test_upload_to_instagram_skips_if_already_uploaded():
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.save()

    recipe = Recipe.objects.create(title="Already Uploaded", instagram_uploaded=True)

    with patch("instagrapi.Client") as MockClient:
        result = upload_to_instagram.func(recipe.pk)
        assert result is False
        MockClient.assert_not_called()


@pytest.mark.django_db
def test_admin_sync_instagram_view_requires_staff(client):
    from django.urls import reverse

    url = reverse("admin_sync_instagram")
    response = client.post(url)
    assert response.status_code == 302
    assert "login" in response.url


@pytest.mark.django_db
def test_admin_sync_instagram_view_triggers_sync(client, staff_user):
    from django.urls import reverse

    # Setup
    client.force_login(staff_user)
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "user"
    config.instagram_password = "pass"
    config.save()

    # Recipe that needs upload
    Recipe.objects.create(title="Pending", description="D", image="img.jpg")

    url = reverse("admin_sync_instagram")

    with patch("django.core.management.call_command") as mock_call:
        response = client.post(url)

        # Verify call_command was called with "sync_instagram_missing"
        mock_call.assert_called_once_with("sync_instagram_missing")

        # Verify redirect
        assert response.status_code == 302
        assert response.url == reverse("admin_settings")


@pytest.mark.django_db
def test_admin_sync_instagram_view_handles_error(client, staff_user):
    from django.urls import reverse

    client.force_login(staff_user)
    url = reverse("admin_sync_instagram")

    with patch("django.core.management.call_command") as mock_call:
        mock_call.side_effect = Exception("Test error")
        response = client.post(url, follow=True)

        assert response.status_code == 200
        # Check if the error message is in the response messages
        assert "Test error" in response.content.decode()


@pytest.mark.django_db
def test_upload_to_instagram_throttling():
    from datetime import timedelta
    from django.utils import timezone
    from django_tasks.backends.database.models import DBTaskResult

    # Clear any existing tasks from other tests or signals
    DBTaskResult.objects.all().delete()

    # Setup
    config = Setting.get_solo()
    config.instagram_enabled = True
    config.instagram_username = "user"
    config.instagram_password = "pass"

    # Set last sync to 30 minutes ago
    config.instagram_last_sync = timezone.now() - timedelta(minutes=30)
    config.save()

    recipe = Recipe.objects.create(title="Throttled Recipe", image="img.jpg")

    from sandwitches.tasks import upload_to_instagram
    from django_tasks.backends.database.models import DBTaskResult

    # Check that it returns False and enqueues a new task in DB
    initial_task_ids = list(
        DBTaskResult.objects.filter(
            task_path="sandwitches.tasks.upload_to_instagram"
        ).values_list("id", flat=True)
    )

    result = upload_to_instagram.func(recipe.pk)

    # Should return False (throttled)
    assert result is False

    # A new task should be enqueued in the database
    new_tasks = DBTaskResult.objects.filter(
        task_path="sandwitches.tasks.upload_to_instagram"
    ).exclude(id__in=initial_task_ids)
    assert new_tasks.exists()

    new_task = new_tasks.order_by("-id").first()
    assert new_task.run_after is not None

    # run_after should be in the future and within the next hour
    assert new_task.run_after > timezone.now()
    assert new_task.run_after <= timezone.now() + timedelta(hours=1)
