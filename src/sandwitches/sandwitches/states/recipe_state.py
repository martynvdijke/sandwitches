import reflex as rx
from typing import TypedDict


class Recipe(TypedDict):
    title: str
    description: str
    ingredients: str
    instructions: str
    image_url: str
    tags: list[str]


class RecipeState(rx.State):
    recipes: list[Recipe] = []
    image_to_add: str = ""
    selected_tags: list[str] = []

    @rx.event
    def load_recipes(self):
        pass

    @rx.var
    def all_tags(self) -> list[str]:
        all_tags = set()
        for recipe in self.recipes:
            for tag in recipe["tags"]:
                all_tags.add(tag.lower())
        return sorted(list(all_tags))

    @rx.var
    def filtered_recipes(self) -> list[Recipe]:
        if not self.selected_tags:
            return self.recipes
        return [
            recipe
            for recipe in self.recipes
            if all(
                (
                    tag in [t.lower() for t in recipe["tags"]]
                    for tag in self.selected_tags
                )
            )
        ]

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
        if not all(form_data.values()):
            yield rx.toast.error("Please fill in all fields.")
            return
        if not self.image_to_add:
            yield rx.toast.error("Please upload an image for the recipe.")
            return
        tags = [tag.strip() for tag in form_data["tags"].split(",") if tag.strip()]
        new_recipe: Recipe = {
            "title": form_data["title"],
            "description": form_data["description"],
            "ingredients": form_data["ingredients"],
            "instructions": form_data["instructions"],
            "image_url": self.image_to_add,
            "tags": tags,
        }
        self.recipes.append(new_recipe)
        self.image_to_add = ""
        yield rx.toast.success(f"Recipe '{form_data['title']}' added successfully!")
        yield rx.clear_selected_files("recipe_image_upload")