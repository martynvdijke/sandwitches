# from gunicorn.http.wsgi import log
import logging

# from django.core.mail import send_mail
from django_tasks import task
from django.contrib.auth import get_user_model

from django.core.mail import EmailMultiAlternatives
from django.conf import settings


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

    for email in emails:
        send_email(recipe_id, email)

    return True

def send_email(recipe_id, email):
    from .models import Recipe
    logging.debug(f"Preparing to send email to: {email}")
    recipe = Recipe.objects.get(pk=recipe_id)  # ty:ignore[unresolved-attribute]
    from_email = getattr(settings, "EMAIL_FROM_ADDRESS")

    full_url = f"https://localhost:8000/{recipe.get_absolute_url()}"

    raw_message = f"""
    Hungry? We just added <strong>{recipe.title}</strong> to our collection.
    
    It's a delicious recipe that you won't want to miss!
    {recipe.description}

    Check out the full recipe, ingredients, and steps here:
    {full_url}

    Happy Cooking!

    The Sandwitches Team
    """
    wrapped_message = textwrap.fill(textwrap.dedent(raw_message), width=70)

    html_content = f"""
    <div style="font-family: 'Helvetica', sans-serif; max-width: 600px; margin: auto; border: 1px solid #eee; padding: 20px;">
        <h2 style="color: #d35400; text-align: center;">New Recipe: {recipe.title} by {recipe.uploaded_by}</h2>
        <div style="text-align: center; margin: 20px 0;">
            <img src="{recipe.image.url}" alt="{recipe.title}" style="width: 100%; border-radius: 8px;">
        </div>
        <p style="font-size: 16px; line-height: 1.5; color: #333;">
            Hungry? We just added <strong>{recipe.title}</strong> to our collection.
            
            It's a delicious recipe that you won't want to miss!
            {recipe.description}

            Click the button below to see how to make it!

            Happy Cooking!

            The Sandwitches Team
        </p>
        <div style="text-align: center; margin-top: 30px;">
            <a href="{full_url}" style="background-color: #e67e22; color: white; padding: 12px 25px; text-decoration: none; border-radius: 5px; font-weight: bold;">VIEW RECIPE</a>
        </div>
    </div>
    """

    msg = EmailMultiAlternatives(
        subject=f"Sandwitches - New Recipe: {recipe.title} by {recipe.uploaded_by}",
        body=wrapped_message,
        from_email=from_email,
        to=[email],
    )
    msg.attach_alternative(html_content, "text/html")
    msg.send()