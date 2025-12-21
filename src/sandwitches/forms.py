from django import forms
from django.contrib.auth import get_user_model
from django.contrib.auth.forms import UserCreationForm
from .models import Recipe

User = get_user_model()


class BaseUserFormMixin:
    """Mixin to handle common password validation and user field processing."""
    
    def clean_passwords(self, cleaned_data):
        p1 = cleaned_data.get("password1")
        p2 = cleaned_data.get("password2")
        if p1 and p2 and p1 != p2:
            raise forms.ValidationError("Passwords do not match.")
        return cleaned_data

    def _set_user_attributes(self, user, data):
        """Helper to apply optional name fields."""
        user.first_name = data.get("first_name", "")
        user.last_name = data.get("last_name", "")
        user.save()
        return user

class AdminSetupForm(forms.ModelForm, BaseUserFormMixin):
    password1 = forms.CharField(widget=forms.PasswordInput, label="Password")
    password2 = forms.CharField(widget=forms.PasswordInput, label="Confirm Password")

    class Meta:
        model = User
        fields = ("username", "first_name", "last_name", "email")

    def clean(self):
        cleaned_data = super().clean()
        return self.clean_passwords(cleaned_data)

    def save(self, commit=True):
        data = self.cleaned_data
        user = User.objects.create_superuser(
            username=data["username"],
            email=data["email"],
            password=data["password1"]
        )
        return self._set_user_attributes(user, data)


class UserSignupForm(UserCreationForm, BaseUserFormMixin):
    """Refactored Regular User Form inheriting from Django's UserCreationForm"""
    
    class Meta(UserCreationForm.Meta):
        model = User
        fields = ("username", "first_name", "last_name", "email")

    def clean(self):
        # UserCreationForm already handles password mismatch, 
        # but we keep this for custom cross-field logic if needed.
        return super().clean()

    def save(self, commit=True):
        user = super().save(commit=False)
        # UserCreationForm's save() handles password hashing automatically
        user.is_superuser = False
        user.is_staff = False
        if commit:
            user.save()
        return user


class RecipeForm(forms.ModelForm):
    class Meta:
        model = Recipe
        fields = ["title", "description", "ingredients", "instructions", "image", "tags"]
        widgets = {
            "tags": forms.TextInput(attrs={"placeholder": "tag1,tag2"}),
        }