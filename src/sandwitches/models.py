from django.db import models
from django.urls import reverse
from django.utils.text import slugify
from .storage import HashedFilenameStorage
from simple_history.models import HistoricalRecords

hashed_storage = HashedFilenameStorage()


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
            while Tag.objects.filter(slug=slug).exclude(pk=self.pk).exists():
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
    image = models.ImageField(
        upload_to="recipes/",  # storage will replace with hashed path
        storage=hashed_storage,
        blank=True,
        null=True,
    )

    # ManyToMany: tags are reusable and shared between recipes
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
            while Recipe.objects.filter(slug=slug).exclude(pk=self.pk).exists():
                slug = f"{base}-{n}"
                n += 1
            self.slug = slug
        super().save(*args, **kwargs)

    def tag_list(self):
        # returns list of tag names
        return list(self.tags.values_list("name", flat=True))

    def set_tags_from_string(self, tag_string):
        """
        Accepts a comma separated string like "tag1, tag2" and attaches existing tags
        or creates new ones as needed. Returns the Tag queryset assigned.
        """
        names = [t.strip() for t in (tag_string or "").split(",") if t.strip()]
        tags = []
        for name in names:
            tag = Tag.objects.filter(name__iexact=name).first()
            if not tag:
                tag = Tag.objects.create(name=name)
            tags.append(tag)
        # replace existing tags with these
        self.tags.set(tags)
        return self.tags.all()

    def get_absolute_url(self):
        return reverse("recipe_detail", kwargs={"pk": self.pk, "slug": self.slug})

    def __str__(self):
        return self.title
