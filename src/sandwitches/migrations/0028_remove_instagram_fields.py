from django.db import migrations


class Migration(migrations.Migration):
    dependencies = [
        ("sandwitches", "0027_setting_log_level"),
    ]

    operations = [
        migrations.RemoveField(
            model_name="historicalrecipe",
            name="instagram_likes_count",
        ),
        migrations.RemoveField(
            model_name="historicalrecipe",
            name="instagram_media_id",
        ),
        migrations.RemoveField(
            model_name="historicalrecipe",
            name="instagram_uploaded",
        ),
        migrations.RemoveField(
            model_name="recipe",
            name="instagram_likes_count",
        ),
        migrations.RemoveField(
            model_name="recipe",
            name="instagram_media_id",
        ),
        migrations.RemoveField(
            model_name="recipe",
            name="instagram_uploaded",
        ),
        migrations.RemoveField(
            model_name="setting",
            name="instagram_enabled",
        ),
        migrations.RemoveField(
            model_name="setting",
            name="instagram_initial_uploaded",
        ),
        migrations.RemoveField(
            model_name="setting",
            name="instagram_last_sync",
        ),
        migrations.RemoveField(
            model_name="setting",
            name="instagram_password",
        ),
        migrations.RemoveField(
            model_name="setting",
            name="instagram_username",
        ),
        migrations.DeleteModel(
            name="InstagramComment",
        ),
    ]
