"""Context processors for sandwitches application."""

from django.conf import settings


def umami(request):
    """Add Umami tracking configuration to template context."""
    return {
        "UMAMI_HOST": settings.UMAMI_HOST,
        "UMAMI_WEBSITE_ID": settings.UMAMI_WEBSITE_ID,
    }


def eink_mode(request):
    """Detect e-ink mode: URL param ?eink=1 > cookie > user theme setting."""
    # Check URL parameter first
    if request.GET.get("eink") == "1":
        return {"eink_mode": True}

    # Check cookie
    if request.COOKIES.get("eink_mode") == "1":
        return {"eink_mode": True}

    # Check authenticated user's theme setting
    if request.user.is_authenticated and request.user.theme == "eink":
        return {"eink_mode": True}

    return {"eink_mode": False}
