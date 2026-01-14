from django.shortcuts import render, get_object_or_404, redirect
from django.urls import reverse
from django.contrib import messages
from django.contrib.auth import login
from django.contrib.auth import get_user_model
from django.contrib.auth.decorators import login_required
from django.contrib.admin.views.decorators import staff_member_required
from django.utils.translation import gettext as _
from .models import Recipe, Rating, Tag
from .forms import (
    RecipeForm,
    AdminSetupForm,
    UserSignupForm,
    RatingForm,
    UserEditForm,
    TagForm,
)
from django.http import HttpResponseBadRequest
from django.conf import settings
from django.http import FileResponse, Http404
from pathlib import Path
import mimetypes
from PIL import Image
from django.db.models import Q, Avg
from django_tasks.backends.database.models import DBTaskResult

from sandwitches import __version__ as sandwitches_version

User = get_user_model()


@staff_member_required
def admin_dashboard(request):
    recipe_count = Recipe.objects.count()  # ty:ignore[unresolved-attribute]
    user_count = User.objects.count()
    tag_count = Tag.objects.count()  # ty:ignore[unresolved-attribute]
    recent_recipes = Recipe.objects.order_by("-created_at")[:5]  # ty:ignore[unresolved-attribute]

    # Data for charts
    from datetime import timedelta
    from django.utils import timezone
    from django.db.models.functions import TruncDate
    from django.db.models import Count

    # Get date range from request or default to last 30 days
    end_date_str = request.GET.get("end_date")
    start_date_str = request.GET.get("start_date")

    today = timezone.now().date()
    try:
        end_date = (
            timezone.datetime.strptime(end_date_str, "%Y-%m-%d").date()
            if end_date_str
            else today
        )
        start_date = (
            timezone.datetime.strptime(start_date_str, "%Y-%m-%d").date()
            if start_date_str
            else end_date - timedelta(days=30)
        )
    except (ValueError, TypeError):
        end_date = today
        start_date = today - timedelta(days=30)

    # Recipes over time
    recipe_data = (
        Recipe.objects.filter(created_at__date__range=(start_date, end_date))  # ty:ignore[unresolved-attribute]
        .annotate(date=TruncDate("created_at"))
        .values("date")
        .annotate(count=Count("id"))
        .order_by("date")
    )

    # Ratings over time
    rating_data = (
        Rating.objects.filter(created_at__date__range=(start_date, end_date))  # ty:ignore[unresolved-attribute]
        .annotate(date=TruncDate("created_at"))
        .values("date")
        .annotate(avg=Avg("score"))
        .order_by("date")
    )

    # Prepare labels and data for JS
    recipe_labels = [d["date"].strftime("%d/%m/%Y") for d in recipe_data]
    recipe_counts = [d["count"] for d in recipe_data]

    rating_labels = [d["date"].strftime("%d/%m/%Y") for d in rating_data]
    rating_avgs = [float(d["avg"]) for d in rating_data]

    return render(
        request,
        "admin/dashboard.html",
        {
            "recipe_count": recipe_count,
            "user_count": user_count,
            "tag_count": tag_count,
            "recent_recipes": recent_recipes,
            "recipe_labels": recipe_labels,
            "recipe_counts": recipe_counts,
            "rating_labels": rating_labels,
            "rating_avgs": rating_avgs,
            "start_date": start_date.strftime("%Y-%m-%d"),
            "end_date": end_date.strftime("%Y-%m-%d"),
            "version": sandwitches_version,
        },
    )


@staff_member_required
def admin_recipe_list(request):
    recipes = (
        Recipe.objects.annotate(avg_rating=Avg("ratings__score"))  # ty:ignore[unresolved-attribute]
        .prefetch_related("tags")
        .all()
    )
    return render(
        request,
        "admin/recipe_list.html",
        {"recipes": recipes, "version": sandwitches_version},
    )


@staff_member_required
def admin_recipe_add(request):
    if request.method == "POST":
        form = RecipeForm(request.POST, request.FILES)
        if form.is_valid():
            recipe = form.save(commit=False)
            recipe.uploaded_by = request.user
            recipe.save()
            form.save_m2m()
            messages.success(request, _("Recipe added successfully."))
            return redirect("admin_recipe_list")
    else:
        form = RecipeForm()
    return render(
        request,
        "admin/recipe_form.html",
        {"form": form, "title": _("Add Recipe"), "version": sandwitches_version},
    )


@staff_member_required
def admin_recipe_edit(request, pk):
    recipe = get_object_or_404(Recipe, pk=pk)
    if request.method == "POST":
        form = RecipeForm(request.POST, request.FILES, instance=recipe)
        if form.is_valid():
            form.save()
            messages.success(request, _("Recipe updated successfully."))
            return redirect("admin_recipe_list")
    else:
        form = RecipeForm(instance=recipe)
    return render(
        request,
        "admin/recipe_form.html",
        {
            "form": form,
            "recipe": recipe,
            "title": _("Edit Recipe"),
            "version": sandwitches_version,
        },
    )


@staff_member_required
def admin_recipe_delete(request, pk):
    recipe = get_object_or_404(Recipe, pk=pk)
    if request.method == "POST":
        recipe.delete()
        messages.success(request, _("Recipe deleted."))
        return redirect("admin_recipe_list")
    return render(
        request,
        "admin/confirm_delete.html",
        {"object": recipe, "type": _("recipe"), "version": sandwitches_version},
    )


@staff_member_required
def admin_recipe_rotate(request, pk):
    recipe = get_object_or_404(Recipe, pk=pk)
    direction = request.GET.get("direction", "cw")
    if not recipe.image:
        messages.error(request, _("No image to rotate."))
        return redirect("admin_recipe_edit", pk=pk)

    try:
        img = Image.open(recipe.image.path)
        if direction == "ccw":
            # Rotate 90 degrees counter-clockwise
            img = img.rotate(90, expand=True)
        else:
            # Rotate 90 degrees clockwise (default)
            img = img.rotate(-90, expand=True)
        img.save(recipe.image.path)
        messages.success(request, _("Image rotated successfully."))
    except Exception as e:
        messages.error(request, _("Error rotating image: ") + str(e))

    return redirect("admin_recipe_edit", pk=pk)


@staff_member_required
def admin_user_list(request):
    users = User.objects.all()
    return render(
        request,
        "admin/user_list.html",
        {"users": users, "version": sandwitches_version},
    )


@staff_member_required
def admin_user_edit(request, pk):
    user_obj = get_object_or_404(User, pk=pk)
    if request.method == "POST":
        form = UserEditForm(request.POST, request.FILES, instance=user_obj)
        if form.is_valid():
            form.save()
            messages.success(request, _("User updated successfully."))
            return redirect("admin_user_list")
    else:
        form = UserEditForm(instance=user_obj)
    return render(
        request,
        "admin/user_form.html",
        {"form": form, "user_obj": user_obj, "version": sandwitches_version},
    )


@staff_member_required
def admin_user_delete(request, pk):
    user_obj = get_object_or_404(User, pk=pk)
    if user_obj == request.user:
        messages.error(request, _("You cannot delete yourself."))
        return redirect("admin_user_list")
    if request.method == "POST":
        user_obj.delete()
        messages.success(request, _("User deleted."))
        return redirect("admin_user_list")
    return render(
        request,
        "admin/confirm_delete.html",
        {"object": user_obj, "type": _("user"), "version": sandwitches_version},
    )


@staff_member_required
def admin_tag_list(request):
    tags = Tag.objects.all()  # ty:ignore[unresolved-attribute]
    return render(
        request,
        "admin/tag_list.html",
        {"tags": tags, "version": sandwitches_version},
    )


@staff_member_required
def admin_tag_add(request):
    if request.method == "POST":
        form = TagForm(request.POST)
        if form.is_valid():
            form.save()
            messages.success(request, _("Tag added successfully."))
            return redirect("admin_tag_list")
    else:
        form = TagForm()
    return render(
        request,
        "admin/tag_form.html",
        {"form": form, "title": _("Add Tag"), "version": sandwitches_version},
    )


@staff_member_required
def admin_tag_edit(request, pk):
    tag = get_object_or_404(Tag, pk=pk)
    if request.method == "POST":
        form = TagForm(request.POST, instance=tag)
        if form.is_valid():
            form.save()
            messages.success(request, _("Tag updated successfully."))
            return redirect("admin_tag_list")
    else:
        form = TagForm(instance=tag)
    return render(
        request,
        "admin/tag_form.html",
        {
            "form": form,
            "tag": tag,
            "title": _("Edit Tag"),
            "version": sandwitches_version,
        },
    )


@staff_member_required
def admin_tag_delete(request, pk):
    tag = get_object_or_404(Tag, pk=pk)
    if request.method == "POST":
        tag.delete()
        messages.success(request, _("Tag deleted."))
        return redirect("admin_tag_list")
    return render(
        request,
        "admin/confirm_delete.html",
        {"object": tag, "type": _("tag"), "version": sandwitches_version},
    )


@staff_member_required
def admin_task_list(request):
    tasks = DBTaskResult.objects.all().order_by("-enqueued_at")[:50]
    return render(
        request,
        "admin/task_list.html",
        {"tasks": tasks, "version": sandwitches_version},
    )


@staff_member_required
def admin_task_detail(request, pk):
    task = get_object_or_404(DBTaskResult, pk=pk)
    return render(
        request,
        "admin/task_detail.html",
        {"task": task, "version": sandwitches_version},
    )


@staff_member_required
def admin_rating_list(request):
    ratings = (
        Rating.objects.select_related("recipe", "user")  # ty:ignore[unresolved-attribute]
        .all()
        .order_by("-updated_at")
    )
    return render(
        request,
        "admin/rating_list.html",
        {"ratings": ratings, "version": sandwitches_version},
    )


@staff_member_required
def admin_rating_delete(request, pk):
    rating = get_object_or_404(Rating, pk=pk)
    if request.method == "POST":
        rating.delete()
        messages.success(request, _("Rating deleted."))
        return redirect("admin_rating_list")
    return render(
        request,
        "admin/confirm_delete.html",
        {"object": rating, "type": _("rating"), "version": sandwitches_version},
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
        initial = (
            {"score": str(user_rating.score), "comment": user_rating.comment}
            if user_rating
            else None
        )
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
            "all_ratings": recipe.ratings.select_related("user").order_by(
                "-created_at"
            ),  # Add all ratings for display
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
        score = form.cleaned_data["score"]
        comment = form.cleaned_data["comment"]
        Rating.objects.update_or_create(  # ty:ignore[unresolved-attribute]
            recipe=recipe,
            user=request.user,
            defaults={"score": score, "comment": comment},
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


@login_required
def favorites(request):
    recipes = request.user.favorites.all().prefetch_related("favorited_by")

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
        return render(
            request,
            "partials/recipe_list.html",
            {"recipes": recipes, "user": request.user},
        )

    # Context for filters - only show options relevant to favorited recipes
    uploaders = User.objects.filter(recipes__in=request.user.favorites.all()).distinct()
    tags = Tag.objects.filter(recipes__in=request.user.favorites.all()).distinct()  # ty:ignore[unresolved-attribute]

    return render(
        request,
        "favorites.html",
        {
            "recipes": recipes,
            "version": sandwitches_version,
            "uploaders": uploaders,
            "tags": tags,
            "selected_tags": request.GET.getlist("tag"),  # Add selected_tags
            "user": request.user,
        },
    )


def index(request):
    if not User.objects.filter(is_superuser=True).exists():
        return redirect("setup")
    recipes = (
        Recipe.objects.all().prefetch_related("favorited_by")  # ty:ignore[unresolved-attribute]
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

    tags = request.GET.getlist("tag")
    if tags:
        recipes = recipes.filter(tags__name__in=tags)

    if request.user.is_authenticated and request.GET.get("favorites") == "on":
        recipes = recipes.filter(pk__in=request.user.favorites.values("pk"))

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
        return render(
            request,
            "partials/recipe_list.html",
            {"recipes": recipes, "user": request.user},
        )

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
            "selected_tags": request.GET.getlist("tag"),
            "user": request.user,  # Pass user to template
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
