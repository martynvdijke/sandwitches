from ninja import NinjaAPI
from .models import Recipe, Tag, Setting
from .utils import (
    parse_ingredient_line,
    scale_ingredient,
    format_scaled_ingredient,
)  # Import utility functions

from ninja import ModelSchema
from ninja import Schema
from django.contrib.auth.models import User
from django.shortcuts import get_object_or_404
from datetime import date
import random
from typing import List, Optional  # Import typing hints

from ninja.security import django_auth

from __init__ import __version__

api = NinjaAPI(version=__version__)


class RecipeSchema(ModelSchema):
    class Meta:
        model = Recipe
        fields = "__all__"


class TagSchema(ModelSchema):
    class Meta:
        model = Tag
        fields = "__all__"


class UserSchema(ModelSchema):
    class Meta:
        model = User
        exclude = ["password", "last_login", "user_permissions"]


class SettingSchema(ModelSchema):
    class Meta:
        model = Setting
        fields = "__all__"


class Error(Schema):
    message: str


class RatingResponseSchema(Schema):
    average: float
    count: int


class ScaledIngredient(Schema):  # New Schema for scaled ingredients
    original_line: str
    scaled_line: str
    quantity: Optional[float]
    unit: Optional[str]
    name: Optional[str]


@api.get("ping")
def ping(request):
    return {"status": "ok", "message": "pong"}


@api.get("v1/settings", response=SettingSchema)
def get_settings(request):
    return Setting.objects.get()


@api.post("v1/settings", auth=django_auth, response={200: SettingSchema, 403: Error})
def update_settings(request, payload: SettingSchema):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to perform this action"}
    settings = Setting.objects.get()
    for attr, value in payload.dict().items():
        setattr(settings, attr, value)
    settings.save()
    return settings


@api.get("v1/me", response={200: UserSchema, 403: Error})
def me(request):
    if not request.user.is_authenticated:
        return 403, {"message": "Please sign in first"}
    return request.user


@api.get("v1/users", auth=django_auth, response=list[UserSchema])
def users(request):
    return User.objects.all()


@api.get("v1/recipes", response=list[RecipeSchema])
def get_recipes(request):
    return Recipe.objects.all()  # ty:ignore[unresolved-attribute]


@api.get("v1/recipes/{recipe_id}", response=RecipeSchema)
def get_recipe(request, recipe_id: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    return recipe


@api.get(
    "v1/recipes/{recipe_id}/scale-ingredients", response=List[ScaledIngredient]
)  # New scaling endpoint
def scale_recipe_ingredients(request, recipe_id: int, target_servings: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)

    current_servings = recipe.servings
    if current_servings is None or current_servings <= 0:
        # Fallback if servings is not set or invalid, assume 1
        current_servings = 1

    ingredients_text = recipe.ingredients
    ingredient_lines = [
        line.strip() for line in ingredients_text.split("\n") if line.strip()
    ]

    scaled_ingredients_output = []
    for line in ingredient_lines:
        parsed = parse_ingredient_line(line)
        scaled = scale_ingredient(parsed, current_servings, target_servings)
        formatted_line = format_scaled_ingredient(scaled)

        scaled_ingredients_output.append(
            ScaledIngredient(
                original_line=line,
                scaled_line=formatted_line,
                quantity=scaled["quantity"],
                unit=scaled["unit"],
                name=scaled["name"],
            )
        )

    return scaled_ingredients_output


@api.get("v1/recipe-of-the-day", response=RecipeSchema)
def get_recipe_of_the_day(request):
    recipes = list(Recipe.objects.all())  # ty:ignore[unresolved-attribute]
    if not recipes:
        return None
    today = date.today()
    random.seed(today.toordinal())
    recipe = random.choice(recipes)
    return recipe


@api.get("v1/recipes/{recipe_id}/rating", response=RatingResponseSchema)
def get_recipe_rating(request, recipe_id: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    return {
        "average": recipe.average_rating(),
        "count": recipe.rating_count(),
    }


@api.get("v1/tags", response=list[TagSchema])
def get_tags(request):
    return Tag.objects.all()  # ty:ignore[unresolved-attribute]


@api.get("v1/tags/{tag_id}", response=TagSchema)
def get_tag(request, tag_id: int):
    tag = get_object_or_404(Tag, id=tag_id)
    return tag
