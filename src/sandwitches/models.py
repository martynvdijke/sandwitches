from django.db import models
from django.utils.text import slugify
from .storage import HashedFilenameStorage
from simple_history.models import HistoricalRecords
from django.contrib.auth.models import AbstractUser
from django.db.models import Avg
from .tasks import email_users
from django.conf import settings
from django.core.validators import MinValueValidator, MaxValueValidator
import logging
from django.urls import reverse
from solo.models import SingletonModel

from imagekit.models import ImageSpecField
from imagekit.processors import ResizeToFill

hashed_storage = HashedFilenameStorage()


class Setting(SingletonModel):
    site_name = models.CharField(max_length=255, default="Sandwitches")
    site_description = models.TextField(blank=True, null=True)
    email = models.EmailField(blank=True, null=True)
    ai_connection_point = models.URLField(blank=True, null=True)
    ai_model = models.CharField(max_length=255, blank=True, null=True)
    ai_api_key = models.CharField(max_length=255, blank=True, null=True)

    def __str__(self):
        return "Site Settings"

    class Meta:
        verbose_name = "Site Settings"


class User(AbstractUser):
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
    favorites = models.ManyToManyField(
        "Recipe", related_name="favorited_by", blank=True
    )

    class Meta:
        verbose_name = "User"
        verbose_name_plural = "Users"

    def __str__(self):
        return self.username


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
