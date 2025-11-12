from django import forms
from django.contrib.auth import get_user_model
from .models import Recipe

User = get_user_model()


class AdminSetupForm(forms.Form):
    username = forms.CharField(max_length=150, label="Username")
    email = forms.EmailField(required=False, label="Email (optional)")
    password1 = forms.CharField(label="Password", widget=forms.PasswordInput)
    password2 = forms.CharField(label="Confirm password", widget=forms.PasswordInput)
    first_name = forms.CharField(max_length=30, required=False, label="First name")
    last_name = forms.CharField(max_length=150, required=False, label="Last name")

    def clean_username(self):
        username = self.cleaned_data.get("username")
        if User.objects.filter(username=username).exists():
            raise forms.ValidationError("A user with that username already exists.")
        return username

    def clean(self):
        cleaned = super().clean()
        p1 = cleaned.get("password1")
        p2 = cleaned.get("password2")
        if p1 and p2 and p1 != p2:
            raise forms.ValidationError("Passwords do not match.")
        return cleaned

    def save(self):
        data = self.cleaned_data
        user = User.objects.create_superuser(
            username=data["username"],
            email=data.get("email") or "",
            password=data["password1"],
        )
        # set optional names
        if data.get("first_name"):
            user.first_name = data["first_name"]
        if data.get("last_name"):
            user.last_name = data["last_name"]
        user.is_staff = True
        user.save()
        return user


class RecipeForm(forms.ModelForm):
    class Meta:
        model = Recipe
        fields = [
            "title",
            "description",
            "ingredients",
            "instructions",
            "image",
            "tags",
        ]
        widgets = {
            "tags": forms.TextInput(attrs={"placeholder": "tag1,tag2"}),
        }
