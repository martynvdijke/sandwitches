import logging
from django.core.mail import send_mail
from django.tasks import task


logger = logging.getLogger(__name__)


@task(takes_context=True, priority=2, queue_name="emails")
def email_users(context, emails, subject, message):
    logger.debug(
        f"Attempt {context.attempt} to send user email. Task result id: {context.task_result.id}."
    )
    return send_mail(
        subject=subject, message=message, from_email=None, recipient_list=emails
    )
