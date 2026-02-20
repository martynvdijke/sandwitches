import logging
from ninja import NinjaAPI
from .models import Recipe, Tag, Setting, Rating, Order, OrderItem, CartItem
from django.contrib.auth import get_user_model
from .utils import (
    parse_ingredient_line,
    scale_ingredient,
    format_scaled_ingredient,
)  # Import utility functions

from ninja import ModelSchema
from ninja import Schema
from django.shortcuts import get_object_or_404
from datetime import date
import random
from typing import List, Optional  # Import typing hints
from django.core.exceptions import ValidationError

from ninja.security import django_auth

from . import __version__

# Get the custom User model
User = get_user_model()

api = NinjaAPI(version=__version__)


class UserPublicSchema(ModelSchema):
    class Meta:
        model = User
        fields = ["username", "first_name", "last_name", "avatar"]


class TagSchema(ModelSchema):
    class Meta:
        model = Tag
        fields = "__all__"


class TagCreateSchema(Schema):
    name: str


class RecipeSchema(ModelSchema):
    favorited_by: List[UserPublicSchema] = []
    tags: List[TagSchema] = []

    class Meta:
        model = Recipe
        fields = "__all__"


class RecipeCreateSchema(Schema):
    title: str
    description: Optional[str] = ""
    ingredients: Optional[str] = ""
    instructions: Optional[str] = ""
    servings: Optional[int] = 1
    price: Optional[float] = None
    tags: Optional[List[str]] = []


class RecipeUpdateSchema(Schema):
    title: Optional[str] = None
    description: Optional[str] = None
    ingredients: Optional[str] = None
    instructions: Optional[str] = None
    servings: Optional[int] = None
    price: Optional[float] = None
    tags: Optional[List[str]] = None


class UserSchema(ModelSchema):
    class Meta:
        model = User
        exclude = ["password", "last_login", "user_permissions", "groups"]


class UserUpdateSchema(Schema):
    first_name: Optional[str] = None
    last_name: Optional[str] = None
    bio: Optional[str] = None
    language: Optional[str] = None
    theme: Optional[str] = None


class SettingSchema(ModelSchema):
    class Meta:
        model = Setting
        fields = "__all__"


class RatingSchema(ModelSchema):
    user: UserPublicSchema

    class Meta:
        model = Rating
        fields = "__all__"


class RatingCreateSchema(Schema):
    score: float
    comment: Optional[str] = ""


class Error(Schema):
    message: str


class RatingResponseSchema(Schema):
    average: float
    count: int


class ScaledIngredient(Schema):
    original_line: str
    scaled_line: str
    quantity: Optional[float]
    unit: Optional[str]
    name: Optional[str]


class OrderSchema(ModelSchema):
    class Meta:
        model = Order
        fields = ["id", "status", "total_price", "created_at", "updated_at"]


class CreateOrderSchema(Schema):
    recipe_id: int


class CartItemSchema(ModelSchema):
    recipe: RecipeSchema

    class Meta:
        model = CartItem
        fields = ["id", "recipe", "quantity", "created_at", "updated_at"]


class CartItemCreateSchema(Schema):
    recipe_id: int
    quantity: Optional[int] = 1


class CartItemUpdateSchema(Schema):
    quantity: int


class OrderItemSchema(ModelSchema):
    recipe: RecipeSchema

    class Meta:
        model = OrderItem
        fields = ["id", "quantity", "price"]


class OrderDetailSchema(ModelSchema):
    items: List[OrderItemSchema]

    class Meta:
        model = Order
        fields = ["id", "status", "total_price", "created_at", "updated_at"]


@api.get("ping")
def ping(request):
    return {"status": "ok", "message": "pong"}


@api.get("v1/settings", response=SettingSchema)
def get_settings(request):
    return Setting.objects.get()  # ty:ignore[unresolved-attribute]


@api.post("v1/settings", auth=django_auth, response={200: SettingSchema, 403: Error})
def update_settings(request, payload: SettingSchema):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to perform this action"}
    settings = Setting.objects.get()  # ty:ignore[unresolved-attribute]
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
    return Recipe.objects.all().prefetch_related("favorited_by")  # ty:ignore[unresolved-attribute]


@api.post("v1/recipes", auth=django_auth, response={201: RecipeSchema, 403: Error})
def create_recipe(request, payload: RecipeCreateSchema):
    data = payload.dict()
    tags = data.pop("tags", [])
    recipe = Recipe.objects.create(**data, uploaded_by=request.user)  # ty:ignore[unresolved-attribute]
    if tags:
        # Assuming tags is a list of tag names
        for tag_name in tags:
            tag, _ = Tag.objects.get_or_create(name=tag_name)  # ty:ignore[unresolved-attribute]
            recipe.tags.add(tag)
    return 201, recipe


@api.get("v1/recipes/{recipe_id}", response=RecipeSchema)
def get_recipe(request, recipe_id: int):
    recipe = get_object_or_404(
        Recipe.objects.prefetch_related("favorited_by"),  # ty:ignore[unresolved-attribute]
        id=recipe_id,
    )
    return recipe


@api.patch(
    "v1/recipes/{recipe_id}", auth=django_auth, response={200: RecipeSchema, 403: Error}
)
def update_recipe(request, recipe_id: int, payload: RecipeUpdateSchema):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    if not request.user.is_staff and recipe.uploaded_by != request.user:
        return 403, {"message": "You are not authorized to edit this recipe"}
    for attr, value in payload.dict(exclude_unset=True).items():
        if attr == "tags":
            if value is not None:
                tags = []
                for tag_name in value:
                    tag, _ = Tag.objects.get_or_create(name=tag_name)  # ty:ignore[unresolved-attribute]
                    tags.append(tag)
                recipe.tags.set(tags)
        else:
            setattr(recipe, attr, value)
    recipe.save()
    return recipe


@api.delete(
    "v1/recipes/{recipe_id}", auth=django_auth, response={204: None, 403: Error}
)
def delete_recipe(request, recipe_id: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    if not request.user.is_staff and recipe.uploaded_by != request.user:
        return 403, {"message": "You are not authorized to delete this recipe"}
    recipe.delete()
    return 204, None


@api.get("v1/recipes/{recipe_id}/scale-ingredients", response=List[ScaledIngredient])
def scale_recipe_ingredients(request, recipe_id: int, target_servings: int):
    recipe = get_object_or_404(Recipe, id=recipe_id)

    # Ensure target_servings is at least 1
    target_servings = max(1, target_servings)

    current_servings = recipe.servings
    if not current_servings or current_servings <= 0:
        current_servings = 1

    ingredient_lines = [
        line.strip() for line in (recipe.ingredients or "").split("\n") if line.strip()
    ]

    scaled_ingredients_output = []
    for line in ingredient_lines:
        try:
            parsed = parse_ingredient_line(line)
            scaled = scale_ingredient(parsed, current_servings, target_servings)
            formatted_line = format_scaled_ingredient(scaled)

            scaled_ingredients_output.append(
                ScaledIngredient(
                    original_line=line,
                    scaled_line=formatted_line,
                    quantity=scaled.get("quantity"),
                    unit=scaled.get("unit"),
                    name=scaled.get("name"),
                )
            )
        except Exception as e:
            # Fallback for lines that fail to parse/scale
            scaled_ingredients_output.append(
                ScaledIngredient(
                    original_line=line,
                    scaled_line=line,
                    quantity=None,
                    unit=None,
                    name=line,
                )
            )
            logging.warning(f"Failed to scale ingredient line '{line}': {e}")

    return scaled_ingredients_output


@api.get("v1/recipe-of-the-day", response=RecipeSchema)
def get_recipe_of_the_day(request):
    recipes = list(Recipe.objects.all().prefetch_related("favorited_by"))  # ty:ignore[unresolved-attribute]
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


@api.post(
    "v1/recipes/{recipe_id}/ratings",
    auth=django_auth,
    response={201: RatingSchema, 403: Error},
)
def create_rating(request, recipe_id: int, payload: RatingCreateSchema):
    recipe = get_object_or_404(Recipe, id=recipe_id)
    data = payload.dict()
    rating, created = Rating.objects.update_or_create(  # ty:ignore[unresolved-attribute]
        recipe=recipe, user=request.user, defaults=data
    )
    return 201, rating


@api.get("v1/tags", response=list[TagSchema])
def get_tags(request):
    return Tag.objects.all()  # ty:ignore[unresolved-attribute]


@api.post("v1/tags", auth=django_auth, response={201: TagSchema, 403: Error})
def create_tag(request, payload: TagCreateSchema):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to create tags"}
    tag = Tag.objects.create(**payload.dict())  # ty:ignore[unresolved-attribute]
    return 201, tag


@api.get("v1/tags/{tag_id}", response=TagSchema)
def get_tag(request, tag_id: int):
    tag = get_object_or_404(Tag, id=tag_id)
    return tag


@api.patch("v1/tags/{tag_id}", auth=django_auth, response={200: TagSchema, 403: Error})
def update_tag(request, tag_id: int, payload: TagSchema):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to edit tags"}
    tag = get_object_or_404(Tag, id=tag_id)
    for attr, value in payload.dict(exclude_unset=True).items():
        if attr != "id":
            setattr(tag, attr, value)
    tag.save()
    return tag


@api.delete("v1/tags/{tag_id}", auth=django_auth, response={204: None, 403: Error})
def delete_tag(request, tag_id: int):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to delete tags"}
    tag = get_object_or_404(Tag, id=tag_id)
    tag.delete()
    return 204, None


@api.get("v1/orders", auth=django_auth, response=list[OrderSchema])
def get_orders(request):
    if request.user.is_staff:
        return Order.objects.all()  # ty:ignore[unresolved-attribute]
    return Order.objects.filter(user=request.user)  # ty:ignore[unresolved-attribute]


@api.post("v1/orders", auth=django_auth, response={201: OrderSchema, 400: Error})
def create_order(request, payload: CreateOrderSchema):
    from .models import OrderItem
    from django.db import transaction
    from .tasks import notify_order_submitted, send_gotify_notification

    recipe = get_object_or_404(Recipe, id=payload.recipe_id)
    try:
        with transaction.atomic():
            order = Order.objects.create(user=request.user)  # ty:ignore[unresolved-attribute]
            OrderItem.objects.create(order=order, recipe=recipe, quantity=1)  # ty:ignore[unresolved-attribute]
            order.total_price = recipe.price
            order.save()

            notify_order_submitted.enqueue(order_id=order.pk)
            send_gotify_notification.enqueue(
                title="New Order Received",
                message=f"Order #{order.pk} by {request.user.username}. Total: {order.total_price}€",
                priority=6,
            )

        return 201, order
    except (ValidationError, ValueError) as e:
        return 400, {"message": str(e)}


@api.get("v1/orders/{order_id}", auth=django_auth, response=OrderDetailSchema)
def get_order(request, order_id: int):
    if request.user.is_staff:
        order = get_object_or_404(Order, id=order_id)
    else:
        order = get_object_or_404(Order, id=order_id, user=request.user)
    return order


@api.patch(
    "v1/orders/{order_id}", auth=django_auth, response={200: OrderSchema, 403: Error}
)
def update_order_status(request, order_id: int, status: str):
    if not request.user.is_staff:
        return 403, {"message": "You are not authorized to update order status"}
    order = get_object_or_404(Order, id=order_id)
    order.status = status
    order.save()
    return order


@api.get("v1/cart", auth=django_auth, response=list[CartItemSchema])
def get_cart(request):
    return CartItem.objects.filter(user=request.user)  # ty:ignore[unresolved-attribute]


@api.post("v1/cart", auth=django_auth, response={201: CartItemSchema, 400: Error})
def add_to_cart_api(request, payload: CartItemCreateSchema):
    recipe = get_object_or_404(Recipe, id=payload.recipe_id)
    cart_item, created = CartItem.objects.get_or_create(  # ty:ignore[unresolved-attribute]
        user=request.user, recipe=recipe, defaults={"quantity": payload.quantity}
    )
    if not created:
        cart_item.quantity += payload.quantity
        cart_item.save()
    return 201, cart_item


@api.patch(
    "v1/cart/{item_id}", auth=django_auth, response={200: CartItemSchema, 403: Error}
)
def update_cart_item(request, item_id: int, payload: CartItemUpdateSchema):
    cart_item = get_object_or_404(CartItem, id=item_id, user=request.user)
    cart_item.quantity = payload.quantity
    cart_item.save()
    return cart_item


@api.delete("v1/cart/{item_id}", auth=django_auth, response={204: None, 403: Error})
def delete_cart_item(request, item_id: int):
    cart_item = get_object_or_404(CartItem, id=item_id, user=request.user)
    cart_item.delete()
    return 204, None
