from django.core.management.base import BaseCommand
from sandwitches.models import Recipe, Setting
from sandwitches.tasks import upload_to_instagram


class Command(BaseCommand):
    help = "Enqueues Instagram upload tasks for recipes that haven't been uploaded yet."

    def handle(self, *args, **options):
        config = Setting.get_solo()
        if not config.instagram_enabled:
            self.stdout.write(
                self.style.WARNING("Instagram integration is disabled in settings.")
            )
            return

        if not config.instagram_username or not config.instagram_password:
            self.stdout.write(
                self.style.ERROR("Instagram credentials missing in settings.")
            )
            return

        # Find recipes that have an image, aren't uploaded, and aren't in progress (no media_id)
        recipes = Recipe.objects.filter(  # ty:ignore[unresolved-attribute]
            instagram_uploaded=False, image__isnull=False
        ).exclude(image="")

        if not recipes.exists():
            self.stdout.write(
                self.style.SUCCESS("No recipes found that need uploading to Instagram.")
            )
            return

        self.stdout.write(
            f"Found {recipes.count()} recipes to enqueue for Instagram upload."
        )

        from datetime import timedelta
        from django.utils import timezone

        for i, recipe in enumerate(recipes):
            self.stdout.write(f"Enqueuing upload for: {recipe.title} (ID: {recipe.pk})")
            upload_to_instagram.using(
                run_after=timezone.now() + timedelta(hours=i)
            ).enqueue(recipe_id=recipe.pk)

        self.stdout.write(
            self.style.SUCCESS(f"Successfully enqueued {recipes.count()} tasks.")
        )
