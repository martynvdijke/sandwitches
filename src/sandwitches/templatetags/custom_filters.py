from django import template

register = template.Library()


@register.filter
def split(value, arg):
    return value.split(arg)


@register.filter
def strip(value):
    if isinstance(value, str):
        return value.strip()
    return value


@register.filter
def strip_lines(value):
    """
    Splits a string by newlines, strips each line, and returns a list of non-empty lines.
    """
    if not value:
        return []
    return [line.strip() for line in value.split("\n") if line.strip()]


@register.filter
def iso8601_duration(minutes):
    """
    Converts minutes to ISO 8601 duration format (e.g., 30 -> PT30M, 90 -> PT1H30M)
    """
    if not minutes:
        return None
    try:
        minutes = int(minutes)
        hours = minutes // 60
        mins = minutes % 60
        duration = "PT"
        if hours > 0:
            duration += f"{hours}H"
        if mins > 0 or hours == 0:
            duration += f"{mins}M"
        return duration
    except (ValueError, TypeError):
        return None
