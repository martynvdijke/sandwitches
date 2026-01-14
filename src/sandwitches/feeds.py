from django.contrib.syndication.views import Feed
from django.urls import reverse_lazy
from .models import Recipe


class LatestRecipesFeed(Feed):
    title = "Sandwitches - Latest Recipes"
    link = reverse_lazy(
        "index"
    )  # This should point to the homepage or a list of recipes
    description = "Updates on the newest recipes added to Sandwitches."

    def items(self):
        return Recipe.objects.order_by("-created_at")[:5]  # ty:ignore[unresolved-attribute]

    def item_title(self, item):
        return item.title

    def item_description(self, item):
        return item.description

    def item_link(self, item):
        return item.get_absolute_url()
