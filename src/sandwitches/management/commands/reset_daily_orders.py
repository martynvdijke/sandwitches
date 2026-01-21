from django.core.management.base import BaseCommand
from sandwitches.models import Recipe


class Command(BaseCommand):
    help = "Resets the daily order count for all recipes. Should be run at midnight."

    def handle(self, *args, **options):
        count = Recipe.objects.update(daily_orders_count=0)  # ty:ignore[unresolved-attribute]
        self.stdout.write(
            self.style.SUCCESS(
                f"Successfully reset daily order count for {count} recipes."
            )
        )
