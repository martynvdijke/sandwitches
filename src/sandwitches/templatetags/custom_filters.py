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
