from django.contrib import admin
from .models import Recipe, Tag, Rating


@admin.register(Recipe)
class RecipeAdmin(admin.ModelAdmin):
    list_display = ("title", "uploaded_by", "created_at")
    readonly_fields = ("created_at", "updated_at")

    def save_model(self, request, obj, form, change):
        # set uploaded_by automatically when creating in admin
        if not change and not obj.uploaded_by:
            obj.uploaded_by = request.user
        super().save_model(request, obj, form, change)


admin.site.register(Tag)
admin.site.register(Rating)
