from django.core.management.base import BaseCommand
from sandwitches.tasks import reset_daily_orders


class Command(BaseCommand):
    help = "Resets the daily order count for all recipes. Should be run at midnight."

    def handle(self, *args, **options):
        task_result = reset_daily_orders.enqueue()
        self.stdout.write(
            self.style.SUCCESS(
                f"Enqueued daily order count reset task (Result ID: {task_result.id})."
            )
        )
