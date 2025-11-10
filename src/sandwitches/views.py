from django.shortcuts import render, get_object_or_404, redirect
from django.urls import reverse
from django.contrib import messages
from django.contrib.auth import login
from django.contrib.auth import get_user_model

from .models import Recipe
from .forms import RecipeForm, AdminSetupForm

User = get_user_model()


def recipe_edit(request, pk):
    recipe = get_object_or_404(Recipe, pk=pk)
    if request.method == "POST":
        form = RecipeForm(request.POST, request.FILES, instance=recipe)
        if form.is_valid():
            form.save()
            return redirect("recipes:admin_list")
    else:
        form = RecipeForm(instance=recipe)
    return render(request, "recipe_form.html", {"form": form, "recipe": recipe})


def recipe_detail(request, slug):
    recipe = get_object_or_404(Recipe, slug=slug)
    return render(request, "detail.html", {"recipe": recipe})


def index(request):
    if not User.objects.filter(is_superuser=True).exists():
        return redirect("setup")
    recipes = Recipe.objects.order_by("-created_at")
    return render(request, "index.html", {"recipes": recipes})


def setup(request):
    """
    First-time setup page: create initial superuser if none exists.
    Visible only while there are no superusers in the DB.
    """
    # do not allow access if a superuser already exists
    if User.objects.filter(is_superuser=True).exists():
        return redirect("index")

    if request.method == "POST":
        form = AdminSetupForm(request.POST)
        if form.is_valid():
            user = form.save()
            # log in the newly created admin
            user.backend = "django.contrib.auth.backends.ModelBackend"
            login(request, user)
            messages.success(request, "Admin account created and signed in.")
            return redirect(reverse("admin:index"))
    else:
        form = AdminSetupForm()

    return render(request, "setup.html", {"form": form})
