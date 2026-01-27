from django.contrib import admin
from django.contrib.auth import get_user_model
from django.contrib.auth.admin import UserAdmin
from .models import Recipe, Tag, Rating, Setting, Order
from django.utils.html import format_html
from import_export import resources
from import_export.admin import ImportExportModelAdmin
from solo.admin import SingletonModelAdmin
from .forms import SettingForm


class RecipeResource(resources.ModelResource):
    class Meta:
        model = Recipe


class TagResource(resources.ModelResource):
    class Meta:
        model = Tag


class RatingResource(resources.ModelResource):
    class Meta:
        model = Rating


class OrderResource(resources.ModelResource):
    class Meta:
        model = Order


User = get_user_model()


@admin.register(Setting)
class SettingAdmin(SingletonModelAdmin):
    form = SettingForm


@admin.register(User)
class CustomUserAdmin(UserAdmin):
    fieldsets = UserAdmin.fieldsets + (
        (None, {"fields": ("avatar", "bio", "language", "favorites")}),
    )


@admin.register(Recipe)
class RecipeAdmin(ImportExportModelAdmin):
    resource_classes = [RecipeResource]
    list_display = (
        "title",
        "uploaded_by",
        "price",
        "created_at",
        "is_highlighted",
        "show_url",
    )
    list_editable = ("is_highlighted",)
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


@admin.register(Tag)
class TagAdmin(ImportExportModelAdmin):
    resource_classes = [TagResource]


@admin.register(Rating)
class RatingAdmin(ImportExportModelAdmin):
    resource_classes = [RatingResource]


@admin.register(Order)
class OrderAdmin(ImportExportModelAdmin):
    resource_classes = [OrderResource]
    list_display = (
        "id",
        "user",
        "recipe",
        "status",
        "completed",
        "total_price",
        "created_at",
    )
    list_filter = ("status", "completed", "created_at")
    search_fields = ("user__username", "recipe__title")
    readonly_fields = ("total_price", "created_at", "updated_at")
