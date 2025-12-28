from django.db import models
from django.utils.text import slugify
from .storage import HashedFilenameStorage
from simple_history.models import HistoricalRecords
from django.contrib.auth import get_user_model
from django.db.models import Avg
from .tasks import email_users

hashed_storage = HashedFilenameStorage()

User = get_user_model()


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
    uploaded_by = models.ForeignKey(
        User,
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
    tags = models.ManyToManyField(Tag, blank=True, related_name="recipes")

    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)
    history = HistoricalRecords()

    class Meta:
        ordering = ("-created_at",)
        verbose_name = "Recipe"
        verbose_name_plural = "Recipes"

    def save(self, *args, **kwargs):
        if not self.slug:
            base = slugify(self.title)[:240]
            slug = base
            n = 1
            while Recipe.objects.filter(slug=slug).exclude(pk=self.pk).exists():  # ty:ignore[unresolved-attribute]
                slug = f"{base}-{n}"
                n += 1
            self.slug = slug

        message = f"New recipe added: {self.title}, go check it out at {self.slug}"
        subject = f"Sandwitches - New Recipe: {self.title} by {self.uploaded_by}"
        email_users.enqueue(subject=subject, message=message)
        super().save(*args, **kwargs)

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

    # def get_absolute_url(self):
    #     return reverse("recipe_detail", kwargs={"pk": self.pk, "slug": self.slug})

    def __str__(self):
        return self.title


class Rating(models.Model):
    recipe = models.ForeignKey(Recipe, related_name="ratings", on_delete=models.CASCADE)
    user = models.ForeignKey(User, related_name="ratings", on_delete=models.CASCADE)
    score = models.PositiveSmallIntegerField(choices=[(i, i) for i in range(1, 6)])
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    class Meta:
        unique_together = ("recipe", "user")
        ordering = ("-updated_at",)

    def __str__(self):
        return f"{self.recipe} — {self.score} by {self.user}"
