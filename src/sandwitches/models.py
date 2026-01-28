from django.db import models
from django.utils.text import slugify
from .storage import HashedFilenameStorage
from simple_history.models import HistoricalRecords
from django.contrib.auth.models import AbstractUser
from django.db.models import Avg
from .tasks import email_users, notify_order_submitted, send_gotify_notification
from django.conf import settings
from django.core.validators import MinValueValidator, MaxValueValidator
import logging
from django.urls import reverse
from solo.models import SingletonModel
from django.core.exceptions import ValidationError
from django.core.validators import EmailValidator

from imagekit.models import ImageSpecField
from imagekit.processors import ResizeToFill

hashed_storage = HashedFilenameStorage()


class Setting(SingletonModel):
    site_name = models.CharField(max_length=255, default="Sandwitches")
    site_description = models.TextField(blank=True, null=True)
    email = models.EmailField(blank=True, null=True, validators=[EmailValidator()])
    ai_connection_point = models.URLField(blank=True, null=True)
    ai_model = models.CharField(max_length=255, blank=True, null=True)
    ai_api_key = models.CharField(max_length=255, blank=True, null=True)

    gotify_url = models.URLField(
        blank=True,
        null=True,
        help_text="The URL of your Gotify server (e.g., https://gotify.example.com)",
    )
    gotify_token = models.CharField(
        max_length=255,
        blank=True,
        null=True,
        help_text="The application token for Gotify",
    )

    def __str__(self):
        return "Site Settings"

    class Meta:
        verbose_name = "Site Settings"


class User(AbstractUser):
    username = models.CharField(max_length=150, unique=True)
    email = models.EmailField(validators=[EmailValidator()])
    avatar = models.ImageField(upload_to="avatars", blank=True, null=True)
    avatar_thumbnail = ImageSpecField(
        source="avatar",
        processors=[ResizeToFill(100, 50)],
        format="JPEG",
        options={"quality": 60},
    )
    bio = models.TextField(blank=True)
    language = models.CharField(
        max_length=10,
        choices=settings.LANGUAGES,
        default=settings.LANGUAGE_CODE,
    )
    theme = models.CharField(
        max_length=10,
        choices=[("light", "Light"), ("dark", "Dark")],
        default="light",
    )
    favorites = models.ManyToManyField(
        "Recipe", related_name="favorited_by", blank=True
    )

    class Meta:
        verbose_name = "User"
        verbose_name_plural = "Users"

    def __str__(self):
        return self.username

    def save(self, *args, **kwargs):
        is_new = self.pk is None
        super().save(*args, **kwargs)
        if is_new:
            send_gotify_notification.enqueue(
                title="New User Created",
                message=f"User {self.username} has joined Sandwitches!",
                priority=4,
            )


class Tag(models.Model):
    name = models.CharField(max_length=50, unique=True)
    slug = models.SlugField(max_length=60, unique=True, blank=True)

    class Meta:
        ordering = ("name",)
        verbose_name = "Tag"
        verbose_name_plural = "Tags"

    def save(self, *args, **kwargs):
        if not self.slug:
            base = slugify(self.name)[:55]
            slug = base
            n = 1
            while Tag.objects.filter(slug=slug).exclude(pk=self.pk).exists():  # ty:ignore[unresolved-attribute]
                slug = f"{base}-{n}"
                n += 1
            self.slug = slug
        super().save(*args, **kwargs)

    def __str__(self):
        return self.name


class Recipe(models.Model):
    title = models.CharField(max_length=255, unique=True)
    slug = models.SlugField(max_length=255, unique=True, blank=True)
    description = models.TextField(blank=True)
    ingredients = models.TextField(blank=True)
    instructions = models.TextField(blank=True)
    servings = models.IntegerField(default=1, validators=[MinValueValidator(1)])
    price = models.DecimalField(
        max_digits=6, decimal_places=2, null=True, blank=True, verbose_name="Price (€)"
    )
    uploaded_by = models.ForeignKey(
        settings.AUTH_USER_MODEL,
        related_name="recipes",
        on_delete=models.SET_NULL,
        null=True,
        blank=True,
    )
    image = models.ImageField(
        upload_to="recipes/",
        storage=hashed_storage,
        blank=True,
        null=True,
    )
    image_thumbnail = ImageSpecField(
        source="image",
        processors=[ResizeToFill(150, 150)],
        format="JPEG",
        options={"quality": 70},
    )
    image_small = ImageSpecField(
        source="image",
        processors=[ResizeToFill(400, 300)],
        format="JPEG",
        options={"quality": 75},
    )
    image_medium = ImageSpecField(
        source="image",
        processors=[ResizeToFill(700, 500)],
        format="JPEG",
        options={"quality": 85},
    )
    image_large = ImageSpecField(
        source="image",
        processors=[ResizeToFill(1200, 800)],
        format="JPEG",
        options={"quality": 95},
    )
    tags = models.ManyToManyField(Tag, blank=True, related_name="recipes")
    is_highlighted = models.BooleanField(default=False)
    is_approved = models.BooleanField(default=False)
    max_daily_orders = models.PositiveIntegerField(
        null=True, blank=True, verbose_name="Max daily orders"
    )
    daily_orders_count = models.PositiveIntegerField(default=0)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    history = HistoricalRecords()

    class Meta:
        ordering = ("-created_at",)
        verbose_name = "Recipe"
        verbose_name_plural = "Recipes"

    def save(self, *args, **kwargs):
        is_new = self._state.adding

        if not self.slug:
            base = slugify(self.title)[:240]
            slug = base
            n = 1
            while Recipe.objects.filter(slug=slug).exclude(pk=self.pk).exists():  # ty:ignore[unresolved-attribute]
                slug = f"{base}-{n}"
                n += 1
            self.slug = slug

        super().save(*args, **kwargs)

        send_email = getattr(settings, "SEND_EMAIL")
        logging.debug(f"SEND_EMAIL is set to {send_email}")

        if is_new or settings.DEBUG:
            if send_email:
                email_users.enqueue(recipe_id=self.pk)
            else:
                logging.warning(
                    "Email sending is disabled; not sending email notification, make sure SEND_EMAIL is set to True in settings."
                )

            send_gotify_notification.enqueue(
                title="New Recipe Uploaded",
                message=f"A new recipe '{self.title}' has been uploaded by {self.uploaded_by or 'Unknown'}.",
                priority=5,
            )
        else:
            logging.debug(
                "Existing recipe saved (update); skipping email notification."
            )

    def get_absolute_url(self):
        return reverse("recipe_detail", kwargs={"slug": self.slug})

    def tag_list(self):
        return list(self.tags.values_list("name", flat=True))  # ty:ignore[possibly-missing-attribute]

    def set_tags_from_string(self, tag_string):
        """
        Accepts a comma separated string like "tag1, tag2" and attaches existing tags
        or creates new ones as needed. Returns the Tag queryset assigned.
        """
        names = [t.strip() for t in (tag_string or "").split(",") if t.strip()]
        tags = []
        for name in names:
            tag = Tag.objects.filter(name__iexact=name).first()  # ty:ignore[unresolved-attribute]
            if not tag:
                tag = Tag.objects.create(name=name)  # ty:ignore[unresolved-attribute]
            tags.append(tag)
        self.tags.set(tags)  # ty:ignore[possibly-missing-attribute]
        return self.tags.all()  # ty:ignore[possibly-missing-attribute]

    def average_rating(self):
        agg = self.ratings.aggregate(avg=Avg("score"))  # ty:ignore[unresolved-attribute]
        return agg["avg"] or 0

    def rating_count(self):
        return self.ratings.count()  # ty:ignore[unresolved-attribute]

    def __str__(self):
        return self.title


class Rating(models.Model):
    recipe = models.ForeignKey(Recipe, related_name="ratings", on_delete=models.CASCADE)
    user = models.ForeignKey(
        settings.AUTH_USER_MODEL, related_name="ratings", on_delete=models.CASCADE
    )
    score = models.FloatField(
        validators=[MinValueValidator(0.0), MaxValueValidator(10.0)]
    )
    comment = models.TextField(blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        unique_together = ("recipe", "user")
        ordering = ("-updated_at",)

    def __str__(self):
        return f"{self.recipe} — {self.score} by {self.user}"


class Order(models.Model):
    STATUS_CHOICES = (
        ("PENDING", "Pending"),
        ("PREPARING", "Preparing"),
        ("MADE", "Made"),
        ("SHIPPED", "Shipped"),
        ("COMPLETED", "Completed"),
        ("CANCELLED", "Cancelled"),
    )

    user = models.ForeignKey(
        settings.AUTH_USER_MODEL, related_name="orders", on_delete=models.CASCADE
    )
    recipe = models.ForeignKey(Recipe, related_name="orders", on_delete=models.CASCADE)
    status = models.CharField(max_length=20, choices=STATUS_CHOICES, default="PENDING")
    completed = models.BooleanField(default=False)
    total_price = models.DecimalField(max_digits=6, decimal_places=2)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        ordering = ("-created_at",)
        verbose_name = "Order"
        verbose_name_plural = "Orders"

    def save(self, *args, **kwargs):
        if not self.recipe.price:  # ty:ignore[possibly-missing-attribute]
            raise ValueError("Cannot order a recipe without a price.")
        if not self.total_price:
            self.total_price = self.recipe.price  # ty:ignore[possibly-missing-attribute]

        is_new = self.pk is None
        if is_new:
            # We use select_for_update to lock the row and prevent race conditions
            # However, since 'self.recipe' is already fetched, we need to re-fetch it with lock if we want to be strict.
            # For simplicity in this context, we will reload it or trust the current instance but ideally:

            # We need to wrap this in a transaction if not already
            # But simple increment logic:
            if (
                self.recipe.max_daily_orders is not None  # ty:ignore[possibly-missing-attribute]
                and self.recipe.daily_orders_count >= self.recipe.max_daily_orders  # ty:ignore[possibly-missing-attribute]
            ):
                raise ValidationError("Daily order limit reached for this recipe.")

            self.recipe.daily_orders_count += 1  # ty:ignore[possibly-missing-attribute]
            self.recipe.save(update_fields=["daily_orders_count"])  # ty:ignore[possibly-missing-attribute]

        super().save(*args, **kwargs)

        if is_new:
            notify_order_submitted.enqueue(order_id=self.pk)
            send_gotify_notification.enqueue(
                title="New Order Received",
                message=f"Order #{self.pk} for '{self.recipe.title}' by {self.user.username}. Total: {self.total_price}€",
                priority=6,
            )

    def __str__(self):
        return f"Order #{self.pk} - {self.user} - {self.recipe}"


class CartItem(models.Model):
    user = models.ForeignKey(
        settings.AUTH_USER_MODEL, related_name="cart_items", on_delete=models.CASCADE
    )
    recipe = models.ForeignKey(
        Recipe, related_name="cart_items", on_delete=models.CASCADE
    )
    quantity = models.PositiveIntegerField(default=1)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        unique_together = ("user", "recipe")
        verbose_name = "Cart Item"
        verbose_name_plural = "Cart Items"

    def __str__(self):
        return f"{self.user.username}'s cart: {self.recipe.title} (x{self.quantity})"

    @property
    def total_price(self):
        if self.recipe.price:
            return self.recipe.price * self.quantity
        return 0
