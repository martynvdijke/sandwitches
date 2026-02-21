"""Context processors for sandwitches application."""

from django.conf import settings


def umami(request):
    """Add Umami tracking configuration to template context."""
    return {
        "UMAMI_HOST": settings.UMAMI_HOST,
        "UMAMI_WEBSITE_ID": settings.UMAMI_WEBSITE_ID,
    }
