"""
URL configuration for sandwitches project.

The `urlpatterns` list routes URLs to views. For more information please see:
    https://docs.djangoproject.com/en/5.2/topics/http/urls/
Examples:
Function views
    1. Add an import:  from my_app import views
    2. Add a URL to urlpatterns:  path('', views.home, name='home')
Class-based views
    1. Add an import:  from other_app.views import Home
    2. Add a URL to urlpatterns:  path('', Home.as_view(), name='home')
Including another URLconf
    1. Import the include() function: from django.urls import include, path
    2. Add a URL to urlpatterns:  path('blog/', include('blog.urls'))
"""

from django.contrib import admin
from django.urls import path, include
from . import views
from .api import api
from django.conf.urls.i18n import i18n_patterns
from .feeds import LatestRecipesFeed  # Import the feed class
from django.contrib.auth.views import LogoutView  # Import LogoutView


import os
import sys


urlpatterns = [
    path("i18n/", include("django.conf.urls.i18n")),
    path("signup/", views.signup, name="signup"),
    path("login/", views.CustomLoginView.as_view(), name="login"),
    path("logout/", LogoutView.as_view(next_page="index"), name="logout"),
    path("profile/", views.user_profile, name="user_profile"),
    path("settings/", views.user_settings, name="user_settings"),
    path("orders/<int:pk>/", views.user_order_detail, name="user_order_detail"),
    path("community/", views.community, name="community"),
    path("admin/", admin.site.urls),
    path("api/", api.urls),
    path("media/<path:file_path>", views.media, name="media"),
    path("favorites/", views.favorites, name="favorites"),
    path("cart/", views.view_cart, name="view_cart"),
    path("cart/add/<int:pk>/", views.add_to_cart, name="add_to_cart"),
    path("cart/remove/<int:pk>/", views.remove_from_cart, name="remove_from_cart"),
    path(
        "cart/update/<int:pk>/", views.update_cart_quantity, name="update_cart_quantity"
    ),
    path("cart/checkout/", views.checkout_cart, name="checkout_cart"),
    path("", views.index, name="index"),
    path("feeds/latest/", LatestRecipesFeed(), name="latest_recipes_feed"),
    path(
        "feeds/latest/", LatestRecipesFeed(), name="latest_recipes_feed"
    ),  # Add this line
]

urlpatterns += i18n_patterns(
    path("recipes/<slug:slug>/", views.recipe_detail, name="recipe_detail"),
    path("setup/", views.setup, name="setup"),
    path("recipes/<int:pk>/rate/", views.recipe_rate, name="recipe_rate"),
    path("recipes/<int:pk>/order/", views.order_recipe, name="order_recipe"),
    path("recipes/<int:pk>/favorite/", views.toggle_favorite, name="toggle_favorite"),
    path("dashboard/", views.admin_dashboard, name="admin_dashboard"),
    path("dashboard/recipes/", views.admin_recipe_list, name="admin_recipe_list"),
    path(
        "dashboard/approvals/",
        views.admin_recipe_approval_list,
        name="admin_recipe_approval_list",
    ),
    path("dashboard/recipes/add/", views.admin_recipe_add, name="admin_recipe_add"),
    path(
        "dashboard/recipes/<int:pk>/edit/",
        views.admin_recipe_edit,
        name="admin_recipe_edit",
    ),
    path(
        "dashboard/recipes/<int:pk>/delete/",
        views.admin_recipe_delete,
        name="admin_recipe_delete",
    ),
    path(
        "dashboard/recipes/<int:pk>/approve/",
        views.admin_recipe_approve,
        name="admin_recipe_approve",
    ),
    path(
        "dashboard/recipes/<int:pk>/rotate/",
        views.admin_recipe_rotate,
        name="admin_recipe_rotate",
    ),
    path("dashboard/users/", views.admin_user_list, name="admin_user_list"),
    path(
        "dashboard/users/<int:pk>/edit/", views.admin_user_edit, name="admin_user_edit"
    ),
    path(
        "dashboard/users/<int:pk>/delete/",
        views.admin_user_delete,
        name="admin_user_delete",
    ),
    path("dashboard/tags/", views.admin_tag_list, name="admin_tag_list"),
    path("dashboard/tags/add/", views.admin_tag_add, name="admin_tag_add"),
    path(
        "dashboard/tags/<int:pk>/edit/",
        views.admin_tag_edit,
        name="admin_tag_edit",
    ),
    path(
        "dashboard/tags/<int:pk>/delete/",
        views.admin_tag_delete,
        name="admin_tag_delete",
    ),
    path("dashboard/tasks/", views.admin_task_list, name="admin_task_list"),
    path(
        "dashboard/tasks/<str:pk>/", views.admin_task_detail, name="admin_task_detail"
    ),
    path("dashboard/ratings/", views.admin_rating_list, name="admin_rating_list"),
    path(
        "dashboard/ratings/<int:pk>/delete/",
        views.admin_rating_delete,
        name="admin_rating_delete",
    ),
    path("dashboard/settings/", views.admin_settings, name="admin_settings"),
    path("dashboard/orders/", views.admin_order_list, name="admin_order_list"),
    path(
        "dashboard/orders/<int:pk>/status/",
        views.admin_order_update_status,
        name="admin_order_update_status",
    ),
    prefix_default_language=True,
)

if "test" not in sys.argv or "PYTEST_VERSION" in os.environ:
    from debug_toolbar.toolbar import debug_toolbar_urls

    urlpatterns = [
        *urlpatterns,
    ] + debug_toolbar_urls()
