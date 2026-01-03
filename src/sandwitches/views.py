from django.shortcuts import render, get_object_or_404, redirect
from django.urls import reverse
from django.contrib import messages
from django.contrib.auth import login
from django.contrib.auth import get_user_model
from django.contrib.auth.decorators import login_required
from django.utils.translation import gettext as _
from .models import Recipe, Rating, Tag
from .forms import RecipeForm, AdminSetupForm, UserSignupForm, RatingForm
from django.http import HttpResponseBadRequest
from django.conf import settings
from django.http import FileResponse, Http404
from pathlib import Path
import mimetypes
from django.db.models import Q, Avg

from sandwitches import __version__ as sandwitches_version

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
    return render(
        request,
        "recipe_form.html",
        {"form": form, "recipe": recipe, "version": sandwitches_version},
    )


def recipe_detail(request, slug):
    recipe = get_object_or_404(Recipe, slug=slug)
    avg = recipe.average_rating()
    count = recipe.rating_count()
    user_rating = None
    rating_form = None
    if request.user.is_authenticated:
        try:
            user_rating = Rating.objects.get(recipe=recipe, user=request.user)  # ty:ignore[unresolved-attribute]
        except Rating.DoesNotExist:  # ty:ignore[unresolved-attribute]
            user_rating = None
        # show form prefilled when possible
        initial = {"score": str(user_rating.score)} if user_rating else None
        rating_form = RatingForm(initial=initial)
    return render(
        request,
        "detail.html",
        {
            "recipe": recipe,
            "avg_rating": avg,
            "rating_count": count,
            "user_rating": user_rating,
            "rating_form": rating_form,
            "version": sandwitches_version,
        },
    )


@login_required
def recipe_rate(request, pk):
    """
    Create or update a rating for the given recipe by the logged-in user.
    """
    recipe = get_object_or_404(Recipe, pk=pk)
    if request.method != "POST":
        return redirect("recipe_detail", slug=recipe.slug)

    form = RatingForm(request.POST)
    if form.is_valid():
        score = float(form.cleaned_data["score"])
        Rating.objects.update_or_create(  # ty:ignore[unresolved-attribute]
            recipe=recipe, user=request.user, defaults={"score": score}
        )
        messages.success(request, _("Your rating has been saved."))
    else:
        messages.error(request, _("Could not save rating."))
    return redirect("recipe_detail", slug=recipe.slug)


@login_required
def toggle_favorite(request, pk):
    recipe = get_object_or_404(Recipe, pk=pk)
    if recipe in request.user.favorites.all():
        request.user.favorites.remove(recipe)
        messages.success(request, _("Recipe removed from favorites."))
    else:
        request.user.favorites.add(recipe)
        messages.success(request, _("Recipe added to favorites."))

    # Redirect to the page where the request came from, or default to recipe detail
    referer = request.META.get("HTTP_REFERER")
    if referer:
        return redirect(referer)
    return redirect("recipe_detail", slug=recipe.slug)


def index(request):
    if not User.objects.filter(is_superuser=True).exists():
        return redirect("setup")
    recipes = (
        Recipe.objects.all()  # ty:ignore[unresolved-attribute]
    )  # Start with all, order later

    # Filtering
    q = request.GET.get("q")
    if q:
        recipes = recipes.filter(
            Q(title__icontains=q) | Q(tags__name__icontains=q)
        ).distinct()

    date_start = request.GET.get("date_start")
    if date_start:
        recipes = recipes.filter(created_at__gte=date_start)

    date_end = request.GET.get("date_end")
    if date_end:
        recipes = recipes.filter(created_at__lte=date_end)

    uploader = request.GET.get("uploader")
    if uploader:
        recipes = recipes.filter(uploaded_by__username=uploader)

    tag = request.GET.get("tag")
    if tag:
        recipes = recipes.filter(tags__name=tag)

    # Sorting
    sort = request.GET.get("sort", "date_desc")
    if sort == "date_asc":
        recipes = recipes.order_by("created_at")
    elif sort == "rating":
        recipes = recipes.annotate(avg_rating=Avg("ratings__score")).order_by(
            "-avg_rating", "-created_at"
        )
    elif sort == "user":
        recipes = recipes.order_by("uploaded_by__username", "-created_at")
    else:  # date_desc or default
        recipes = recipes.order_by("-created_at")

    if request.headers.get("HX-Request"):
        return render(request, "partials/recipe_list.html", {"recipes": recipes})

    # Context for filters
    uploaders = User.objects.filter(recipes__isnull=False).distinct()
    tags = Tag.objects.all()  # ty:ignore[unresolved-attribute]

    return render(
        request,
        "index.html",
        {
            "recipes": recipes,
            "version": sandwitches_version,
            "uploaders": uploaders,
            "tags": tags,
        },
    )


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
            user.backend = "django.contrib.auth.backends.ModelBackend"
            login(request, user)
            messages.success(request, _("Admin account created and signed in."))
            return redirect(reverse("admin:index"))
    else:
        form = AdminSetupForm()

    return render(request, "setup.html", {"form": form, "version": sandwitches_version})


def signup(request):
    """
    User signup page: create new regular user accounts.
    """
    if request.method == "POST":
        form = UserSignupForm(request.POST, request.FILES)
        if form.is_valid():
            user = form.save()
            # log in the newly created user
            user.backend = "django.contrib.auth.backends.ModelBackend"
            login(request, user)
            messages.success(request, _("Account created and signed in."))
            return redirect("index")
    else:
        form = UserSignupForm()

    return render(
        request, "signup.html", {"form": form, "version": sandwitches_version}
    )


def media(request, file_path=None):
    media_root = getattr(settings, "MEDIA_ROOT", None)
    if not media_root:
        return HttpResponseBadRequest("Invalid Media Root Configuration")
    if not file_path:
        return HttpResponseBadRequest("Invalid File Path")

    base_path = Path(media_root).resolve()
    full_path = base_path.joinpath(file_path).resolve()
    if base_path not in full_path.parents:
        return HttpResponseBadRequest("Access Denied")

    if not full_path.exists() or not full_path.is_file():
        raise Http404("File not found")

    content_type, encoding = mimetypes.guess_type(full_path)
    if not content_type or not content_type.startswith("image/"):
        return HttpResponseBadRequest("Access Denied: Only image files are allowed.")

    response = FileResponse(open(full_path, "rb"), as_attachment=True)
    return response
