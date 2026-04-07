from django.apps import AppConfig
import logging


class SandwitchesConfig(AppConfig):
    default_auto_field = "django.db.models.BigAutoField"
    name = "sandwitches"

    def ready(self):
        # Apply the logging level from settings at startup
        try:
            from .models import Setting
            from .utils import set_logging_level
            from django.db import connection

            # Check if we are running in a context where database is available
            import sys

            if "manage.py" in sys.argv and any(
                arg in sys.argv
                for arg in ["migrate", "makemigrations", "collectstatic", "test"]
            ):
                return

            # Check if table exists
            if "sandwitches_setting" not in connection.introspection.table_names():
                return

            config = Setting.get_solo()
            if config.log_level:
                set_logging_level(config.log_level)
                logger = logging.getLogger("sandwitches")
                logger.info(f"Logging level set to {config.log_level} at startup.")
        except Exception:
            # Database might not be ready or table might not exist
            pass
