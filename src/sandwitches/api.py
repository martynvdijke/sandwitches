from ninja import NinjaAPI
from .models import Recipe, Tag

from ninja import ModelSchema
from ninja import Schema
from django.contrib.auth.models import User
from django.shortcuts import get_object_or_404
from datetime import date
import random

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


class Error(Schema):
    message: str


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


@api.get("v1/recipe-of-the-day", response=RecipeSchema)
def get_recipe_of_the_day(request):
    recipes = list(Recipe.objects.all())  # ty:ignore[unresolved-attribute]
    if not recipes:
        return None
    today = date.today()
    random.seed(today.toordinal())
    recipe = random.choice(recipes)
    return recipe


@api.get("v1/tags", response=list[TagSchema])
def get_tags(request):
    return Tag.objects.all()  # ty:ignore[unresolved-attribute]


@api.get("v1/tags/{tag_id}", response=TagSchema)
def get_tag(request, tag_id: int):
    tag = get_object_or_404(Tag, id=tag_id)
    return tag
