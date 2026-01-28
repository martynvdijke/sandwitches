import logging
import requests

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


@task(priority=5)
def reset_daily_orders():
    from .models import Recipe

    count = Recipe.objects.update(daily_orders_count=0)  # ty:ignore[unresolved-attribute]
    logging.info(f"Successfully reset daily order count for {count} recipes.")
    return count


@task(priority=2, queue_name="emails")
def notify_order_submitted(order_id):
    from .models import Order

    try:
        order = Order.objects.select_related("user", "recipe").get(pk=order_id)  # ty:ignore[unresolved-attribute]
    except Order.DoesNotExist:  # ty:ignore[unresolved-attribute]
        logging.warning(f"Order {order_id} not found. Skipping notification.")
        return

    user = order.user
    if not user.email:
        logging.warning(f"User {user.username} has no email. Skipping notification.")
        return

    recipe = order.recipe
    subject = _("Order Confirmation: %(recipe_title)s") % {"recipe_title": recipe.title}
    from_email = getattr(settings, "EMAIL_FROM_ADDRESS")

    context_data = {
        "user_name": user.get_full_name() or user.username,
        "recipe_title": recipe.title,
        "order_id": order.id,
        "total_price": order.total_price,
    }

    text_content = (
        _(
            "Hello %(user_name)s,\n\n"
            "Your order for %(recipe_title)s has been successfully submitted!\n"
            "Order ID: %(order_id)s\n"
            "Total Price: %(total_price)s\n\n"
            "Thank you for ordering with Sandwitches.\n"
        )
        % context_data
    )

    html_content = (
        _(
            "<div style='font-family: sans-serif;'>"
            "<h2>Order Confirmation</h2>"
            "<p>Hello <strong>%(user_name)s</strong>,</p>"
            "<p>Your order for <strong>%(recipe_title)s</strong> has been successfully submitted!</p>"
            "<ul>"
            "<li>Order ID: %(order_id)s</li>"
            "<li>Total Price: %(total_price)s</li>"
            "</ul>"
            "<p>Thank you for ordering with Sandwitches.</p>"
            "</div>"
        )
        % context_data
    )

    msg = EmailMultiAlternatives(
        subject=subject,
        body=text_content,
        from_email=from_email,
        to=[user.email],
    )
    msg.attach_alternative(html_content, "text/html")
    msg.send()

    logging.info(f"Order confirmation email sent to {user.email} for order {order.id}")


@task(priority=1, queue_name="emails")
def send_gotify_notification(title, message, priority=5):
    from .models import Setting

    config = Setting.get_solo()
    url = config.gotify_url
    token = config.gotify_token

    if not url or not token:
        logging.debug("Gotify URL or Token not configured. Skipping notification.")
        return False

    try:
        response = requests.post(
            f"{url.rstrip('/')}/message?token={token}",
            json={
                "title": title,
                "message": message,
                "priority": priority,
            },
            timeout=10,
        )
        response.raise_for_status()
        logging.info(f"Gotify notification sent: {title}")
        return True
    except Exception as e:
        logging.error(f"Failed to send Gotify notification: {e}")
        return False


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
