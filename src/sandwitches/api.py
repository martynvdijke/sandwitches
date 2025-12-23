from ninja import NinjaAPI
from .models import Recipe

from ninja import ModelSchema
from ninja import Schema
from django.contrib.auth.models import User
from django.shortcuts import get_object_or_404
from datetime import date
import random


api = NinjaAPI()


class RecipeSchema(ModelSchema):
    class Meta:
        model = Recipe
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


# TODO: enable recipe creation via API
# @api.post("v1/recipe", auth=django_auth, response=RecipeSchema)
# def create_recipe(request, payload: RecipeSchema):
#     recipe = Recipe.objects.create(**payload.dict())
#     return recipe


@api.get("v1/recipes", response=list[RecipeSchema])
def get_recipes(request):
    recipes = Recipe.objects.all()
    return recipes


@api.get("v1/recipes/{recipe_id}", response=RecipeSchema)
def get_recipe(request, recipe_id: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    return recipe


@api.get("v1/recipe-of-the-day", response=RecipeSchema)
def get_recipe_of_the_day(request):
    recipes = list(Recipe.objects.all())
    if not recipes:
        return None
    today = date.today()
    random.seed(today.toordinal())
    recipe = random.choice(recipes)
    return recipe
