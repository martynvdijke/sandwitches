import reflex as rx
from typing import TypedDict, Optional
import json
from sqlalchemy import text

from ..models import Recipe

from sqlmodel import select

class RecipeState(rx.State):
    recipes: list[Recipe] = []
    image_to_add: str = ""
    selected_tags: list[str] = []
    show_edit_modal: bool = False
    recipe_to_edit: Optional[Recipe] = None
    image_for_edit: str = ""
    
    @rx.event
    def load_recipes(self):
        with rx.session() as session:
            self.recipes = session.exec(Recipe.select()).all()
            for recipe in self.recipes:
                recipe.tagstest = [tag.strip() for tag in recipe.tags.split(",")]
            print(self.recipes)
                
    @rx.var
    def selected_recipe(self) -> Optional[Recipe]:
        """Get the selected recipe from the page params."""
        recipe_title = self.router.page.params.get("title", "no-recipe").replace(
            "-", " "
        )
        print(self.recipes)
        return next(
            (
                recipe
                for recipe in self.recipes
                if recipe.title.lower() == recipe_title.lower()
            ),
            None,
        )

    @rx.event
    def start_editing(self, recipe: Recipe):
        self.recipe_to_edit = recipe
        self.image_for_edit = recipe["image_url"]
        self.show_edit_modal = True

    @rx.event
    def cancel_editing(self):
        self.show_edit_modal = False
        self.recipe_to_edit = None
        self.image_for_edit = ""

    @rx.var
    def all_tags(self) -> list[str]:
        all_tags = set()
        for recipe in self.recipes:
            all_tags.update(recipe.tags)
        print(all_tags)
        return sorted(list(all_tags))

    @rx.var
    def filtered_recipes(self) -> list[Recipe]:
        if not self.selected_tags:
            return self.recipes
        print(self.selected_tags)
        test = [
            recipe
            for recipe in self.recipes
            if 
                (
                    tag in [t.strip().lower() for t in recipe.tags.split(",")]
                    for tag in self.selected_tags
                )
            
        ]
        print("filtered recipes:")
        print(test)
        return test
    @rx.event
    def toggle_tag(self, tag: str):
        tag = tag.lower()
        if tag in self.selected_tags:
            self.selected_tags.remove(tag)
        else:
            self.selected_tags.append(tag)

    @rx.event
    async def handle_upload(self, files: list[rx.UploadFile]):
        if not files:
            yield rx.toast.error("Please select an image to upload.")
            return
        for file in files:
            upload_data = await file.read()
            outfile = f"assets/{file.name}"
            with open(outfile, "wb") as file_object:
                file_object.write(upload_data)
            self.image_to_add = file.name
        yield rx.toast.success(f"Successfully uploaded {self.image_to_add}")

    @rx.event
    def add_recipe(self, form_data: dict):
        # if not all(form_data.values()):
        #     yield rx.toast.error("Please fill in all fields.")
        #     return
        # if not self.image_to_add:
        #     yield rx.toast.error("Please upload an image for the recipe.")
        #     return
        new_recipe: Recipe = {
            "title": form_data["title"],
            "description": form_data["description"],
            "ingredients": form_data["ingredients"],
            "instructions": form_data["instructions"],
            "image_url": self.image_to_add,
            "tags": form_data["tags"],
        }
        with rx.session() as session:
            db_recipe = Recipe(**new_recipe)
            session.add(db_recipe)
            session.commit()
            session.refresh(db_recipe)
        yield rx.toast.success(f"Recipe '{form_data['title']}' added successfully!")
        yield rx.clear_selected_files("recipe_image_upload")
        
    @rx.event
    def update_recipe(self, form_data: dict):
        # if not all(form_data.values()):
        #     yield rx.toast.error("Please fill in all fields.")
        #     return
        # if not self.image_to_add:
        #     yield rx.toast.error("Please upload an image for the recipe.")
        #     return
        new_recipe: Recipe = {
            "title": form_data["title"],
            "description": form_data["description"],
            "ingredients": form_data["ingredients"],
            "instructions": form_data["instructions"],
            "image_url": self.image_to_add,
            "tags": form_data["tags"],
        }
        with rx.session() as session:
            db_recipe = Recipe(**new_recipe)
            session.add(new_recipe)
            session.commit()
            session.refresh(db_recipe)
        yield rx.toast.success(f"Recipe '{form_data['title']}' added successfully!")
        yield rx.clear_selected_files("recipe_image_upload")