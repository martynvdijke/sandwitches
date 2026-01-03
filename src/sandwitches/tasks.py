# from gunicorn.http.wsgi import log
import logging

# from django.core.mail import send_mail
from django_tasks import task
from django.contrib.auth import get_user_model

from django.core.mail import EmailMultiAlternatives
from django.conf import settings
from django.utils.translation import gettext as _


import textwrap


@task(takes_context=True, priority=2, queue_name="emails")
def email_users(context, recipe_id):
    logging.debug(
        f"Attempt {context.attempt} to send users an email. Task result id: {context.task_result.id}."
    )

    User = get_user_model()
    emails = list(
        User.objects.exclude(email__isnull=True)
        .exclude(email="")
        .values_list("email", flat=True)
    )

    if not emails:
        logging.warning("No users with valid emails found.")
        return 0

    send_emails(recipe_id, emails)

    return True


def send_emails(recipe_id, emails):
    from .models import Recipe

    logging.debug(f"Preparing to send email to: {emails}")
    recipe = Recipe.objects.get(pk=recipe_id)  # ty:ignore[unresolved-attribute]
    from_email = getattr(settings, "EMAIL_FROM_ADDRESS")

    recipe_slug = recipe.get_absolute_url()
    base_url = (
        settings.CSRF_TRUSTED_ORIGINS[0]
        if settings.CSRF_TRUSTED_ORIGINS
        else "http://localhost"
    ).rstrip("/")

    raw_message_fmt = _("""
    Hungry? We just added <strong>%(title)s</strong> to our collection.
    
    It's a delicious recipe that you won't want to miss!
    %(description)s

    Check out the full recipe, ingredients, and steps here:
    %(url)s

    Happy Cooking!

    The Sandwitches Team
    """)

    context_data = {
        "title": recipe.title,
        "uploaded_by": recipe.uploaded_by,
        "description": recipe.description,
        "url": f"{base_url}{recipe_slug}",
        "image_url": f"{base_url}{recipe.image.url}" if recipe.image else "",
    }

    wrapped_message = textwrap.fill(
        textwrap.dedent(raw_message_fmt) % context_data, width=70
    )

    html_content_fmt = _("""
    <div style="font-family: 'Helvetica', sans-serif; max-width: 600px; margin: auto; border: 1px solid #eee; padding: 20px;">
        <h2 style="color: #d35400; text-align: center;">New Recipe: %(title)s by %(uploaded_by)s</h2>
        <div style="text-align: center; margin: 20px 0;">
            <img src="%(image_url)s" alt="%(title)s" style="width: 100%%; border-radius: 8px;">
        </div>
        <p style="font-size: 16px; line-height: 1.5; color: #333;">
            Hungry? We just added <strong>%(title)s</strong> to our collection.
            <br>
            It's a delicious recipe that you won't want to miss!
            <br>
            %(description)s
            <br>
            Check out the full recipe, ingredients, and steps here:
            Click the button below to see how to make it!
            <br>
            Happy Cooking!
            <br>
            The Sandwitches Team
        </p>
        <div style="text-align: center; margin-top: 30px;">
            <a href="%(url)s" style="background-color: #e67e22; color: white; padding: 12px 25px; text-decoration: none; border-radius: 5px; font-weight: bold;">VIEW RECIPE</a>
        </div>
    </div>
    """)

    html_content = html_content_fmt % context_data

    subject = _("Sandwitches - New Recipe: %(title)s by %(uploaded_by)s") % context_data
    for email in emails:
        msg = EmailMultiAlternatives(
            subject=subject,
            body=wrapped_message,
            from_email=from_email,
            to=[email],
        )
        msg.attach_alternative(html_content, "text/html")
        msg.send()
