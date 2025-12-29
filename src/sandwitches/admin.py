from django.contrib import admin
from .models import Recipe, Tag, Rating
from django.utils.html import format_html


@admin.register(Recipe)
class RecipeAdmin(admin.ModelAdmin):
    list_display = ("title", "uploaded_by", "created_at", "show_url")
    readonly_fields = ("created_at", "updated_at")

    def save_model(self, request, obj, form, change):
        # set uploaded_by automatically when creating in admin
        if not change and not obj.uploaded_by:
            obj.uploaded_by = request.user
        super().save_model(request, obj, form, change)

    def show_url(self, obj):
        url = obj.get_absolute_url()
        return format_html("<a href='{url}'>{url}</a>", url=url)

    show_url.short_description = "Recipe Link"  # ty:ignore[unresolved-attribute]


admin.site.register(Tag)
admin.site.register(Rating)
