from django import forms
from .models import Recipe

class RecipeForm(forms.ModelForm):
    class Meta:
        model = Recipe
        fields = ['title', 'description', 'ingredients', 'instructions', 'image', 'tags']
        widgets = {
            'tags': forms.TextInput(attrs={'placeholder': 'tag1,tag2'}),
        }