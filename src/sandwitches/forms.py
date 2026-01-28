from django import forms
from django.contrib.auth import get_user_model
from django.contrib.auth.forms import UserCreationForm
from django.utils.translation import gettext_lazy as _
from .models import Recipe, Tag, Setting

User = get_user_model()


class BaseUserFormMixin:
    """Mixin to handle common password validation and user field processing."""

    def clean_passwords(self, cleaned_data):
        p1 = cleaned_data.get("password1")
        p2 = cleaned_data.get("password2")
        if p1 and p2 and p1 != p2:
            raise forms.ValidationError(_("Passwords do not match."))
        return cleaned_data

    def _set_user_attributes(self, user, data):
        """Helper to apply optional name fields."""
        user.first_name = data.get("first_name", "")
        user.last_name = data.get("last_name", "")
        user.save()
        return user


class AdminSetupForm(forms.ModelForm, BaseUserFormMixin):
    password1 = forms.CharField(widget=forms.PasswordInput, label=_("Password"))
    password2 = forms.CharField(widget=forms.PasswordInput, label=_("Confirm Password"))

    class Meta:
        model = User
        fields = ("username", "first_name", "last_name", "email")

    def clean(self):
        cleaned_data = super().clean()
        return self.clean_passwords(cleaned_data)

    def save(self, commit=True):
        data = self.cleaned_data
        user = User.objects.create_superuser(
            username=data["username"], email=data["email"], password=data["password1"]
        )
        return self._set_user_attributes(user, data)


class UserSignupForm(UserCreationForm, BaseUserFormMixin):
    avatar = forms.ImageField(label=_("Profile Image"), required=False)
    bio = forms.CharField(
        widget=forms.Textarea(attrs={"rows": 3}), label=_("Bio"), required=False
    )

    class Meta(UserCreationForm.Meta):
        model = User
        fields = (
            "username",
            "first_name",
            "last_name",
            "email",
            "avatar",
            "bio",
        )

    def clean(self):
        return super().clean()

    def save(self, commit=True):
        user = super().save(commit=False)
        user.is_superuser = False
        user.is_staff = False
        # Explicitly save the extra fields if they aren't automatically handled by ModelForm save (they should be if in Meta.fields)
        user.avatar = self.cleaned_data["avatar"]
        user.bio = self.cleaned_data["bio"]
        if commit:
            user.save()
        return user


class UserProfileForm(forms.ModelForm):
    image_data = forms.CharField(widget=forms.HiddenInput(), required=False)

    class Meta:
        model = User
        fields = (
            "first_name",
            "last_name",
            "email",
            "avatar",
            "bio",
        )

    def save(self, commit=True):
        user = super().save(commit=False)
        image_data = self.cleaned_data.get("image_data")
        if image_data and image_data.startswith("data:image"):
            import base64
            from django.core.files.base import ContentFile

            format, imgstr = image_data.split(";base64,")
            ext = format.split("/")[-1]
            data = ContentFile(base64.b64decode(imgstr), name=f"avatar.{ext}")
            user.avatar = data
        if commit:
            user.save()
        return user


class UserSettingsForm(forms.ModelForm):
    class Meta:
        model = User
        fields = ("language", "theme")
        labels = {
            "language": _("Preferred Language"),
            "theme": _("Preferred Theme"),
        }


class UserEditForm(forms.ModelForm):
    image_data = forms.CharField(widget=forms.HiddenInput(), required=False)

    class Meta:
        model = User
        fields = (
            "username",
            "first_name",
            "last_name",
            "email",
            "is_staff",
            "is_active",
            "language",
            "avatar",
            "bio",
        )

    def save(self, commit=True):
        user = super().save(commit=False)
        image_data = self.cleaned_data.get("image_data")
        if image_data and image_data.startswith("data:image"):
            import base64
            from django.core.files.base import ContentFile

            format, imgstr = image_data.split(";base64,")
            ext = format.split("/")[-1]
            data = ContentFile(base64.b64decode(imgstr), name=f"avatar.{ext}")
            user.avatar = data
        if commit:
            user.save()
        return user


class TagForm(forms.ModelForm):
    class Meta:
        model = Tag
        fields = ("name",)


class RecipeForm(forms.ModelForm):
    tags_string = forms.CharField(
        required=False,
        label=_("Tags (comma separated)"),
        widget=forms.TextInput(attrs={"placeholder": _("e.g. spicy, vegan, quick")}),
    )
    rotation = forms.IntegerField(widget=forms.HiddenInput(), initial=0, required=False)
    image_data = forms.CharField(widget=forms.HiddenInput(), required=False)

    class Meta:
        model = Recipe
        fields = [
            "title",
            "image",
            "uploaded_by",
            "description",
            "ingredients",
            "instructions",
            "price",
            "is_highlighted",
            "is_approved",
            "max_daily_orders",
        ]
        widgets = {
            "image": forms.FileInput(),
        }

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        if self.instance.pk:
            self.fields["tags_string"].initial = ", ".join(
                self.instance.tags.values_list("name", flat=True)
            )

    def save(self, commit=True):
        recipe = super().save(commit=False)

        # Handle base64 image data from cropper
        image_data = self.cleaned_data.get("image_data")
        if image_data and image_data.startswith("data:image"):
            import base64
            from django.core.files.base import ContentFile

            format, imgstr = image_data.split(";base64,")
            ext = format.split("/")[-1]
            data = ContentFile(base64.b64decode(imgstr), name=f"recipe_image.{ext}")
            recipe.image = data

        # Handle rotation if an image exists and rotation is requested (fallback for simple rotation)
        rotation = self.cleaned_data.get("rotation", 0)
        if rotation != 0 and recipe.image and not image_data:
            try:
                from PIL import Image as PILImage

                img = PILImage.open(recipe.image.path)
                # PIL rotates counter-clockwise by default, our 'rotation' is clockwise
                img = img.rotate(-rotation, expand=True)
                img.save(recipe.image.path)
            except Exception as e:
                print(f"Error rotating image: {e}")

        if commit:
            recipe.save()
            recipe.set_tags_from_string(self.cleaned_data.get("tags_string", ""))
        else:
            self.save_m2m = lambda: recipe.set_tags_from_string(
                self.cleaned_data.get("tags_string", "")
            )
        return recipe


class UserRecipeSubmissionForm(forms.ModelForm):
    tags_string = forms.CharField(
        required=False,
        label=_("Tags (comma separated)"),
        widget=forms.TextInput(attrs={"placeholder": _("e.g. spicy, vegan, quick")}),
    )
    image_data = forms.CharField(widget=forms.HiddenInput(), required=False)

    class Meta:
        model = Recipe
        fields = [
            "title",
            "image",
            "description",
            "ingredients",
            "instructions",
            "price",
            "servings",
        ]
        widgets = {
            "image": forms.FileInput(),
        }

    def save(self, commit=True):
        recipe = super().save(commit=False)

        # Handle base64 image data from cropper
        image_data = self.cleaned_data.get("image_data")
        if image_data and image_data.startswith("data:image"):
            import base64
            from django.core.files.base import ContentFile

            format, imgstr = image_data.split(";base64,")
            ext = format.split("/")[-1]
            data = ContentFile(base64.b64decode(imgstr), name=f"recipe_image.{ext}")
            recipe.image = data

        if commit:
            recipe.save()
            recipe.set_tags_from_string(self.cleaned_data.get("tags_string", ""))
        else:
            self.save_m2m = lambda: recipe.set_tags_from_string(
                self.cleaned_data.get("tags_string", "")
            )
        return recipe


class RatingForm(forms.Form):
    """Form for rating recipes (0-10) with an optional comment."""

    score = forms.FloatField(
        min_value=0.0,
        max_value=10.0,
        widget=forms.NumberInput(
            attrs={"step": "0.1", "min": "0", "max": "10", "class": "slider"}
        ),
        label=_("Your rating"),
    )
    comment = forms.CharField(
        widget=forms.Textarea(attrs={"rows": 2}),
        label=_("Comment (optional)"),
        required=False,
    )


class SettingForm(forms.ModelForm):
    class Meta:
        model = Setting
        fields = [
            "site_name",
            "site_description",
            "email",
            "ai_connection_point",
            "ai_model",
            "ai_api_key",
            "gotify_url",
            "gotify_token",
        ]
        widgets = {
            "ai_api_key": forms.PasswordInput(render_value=True),
            "gotify_token": forms.PasswordInput(render_value=True),
        }
