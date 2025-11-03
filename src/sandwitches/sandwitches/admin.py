from django.contrib import admin
from .models import Recipe, Tag


admin.site.register(Recipe)
admin.site.register(Tag)
