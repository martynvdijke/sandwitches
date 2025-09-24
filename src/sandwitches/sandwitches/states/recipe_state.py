import reflex as rx
from typing import TypedDict, Optional


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
        if not self.recipes:
            self.recipes = [
                {
                    "title": "Classic Club Sandwich",
                    "description": "A timeless multi-layered sandwich, perfect for a hearty lunch.",
                    "ingredients": "Toasted bread, Chicken or Turkey, Bacon, Lettuce, Tomato, Mayonnaise",
                    "instructions": "1. Toast the bread. 2. Spread mayonnaise on one side of each slice. 3. Layer chicken/turkey, bacon, lettuce, and tomato. 4. Cut into quarters and serve.",
                    "image_url": "https://www.lekkerensimpel.com/wp-content/uploads/2023/04/588A5318.jpg",
                    "tags": ["classic", "chicken", "lunch"],
                },
                {
                    "title": "Ultimate BLT Sandwich",
                    "description": "The perfect combination of bacon, lettuce, and tomato.",
                    "ingredients": "Bacon, Lettuce, Tomato, Bread, Mayonnaise",
                    "instructions": "1. Cook bacon until crispy. 2. Toast bread and spread with mayonnaise. 3. Layer with bacon, lettuce, and tomato slices. 4. Slice in half and enjoy.",
                    "image_url": "https://www.lekkerensimpel.com/wp-content/uploads/2023/04/588A5318.jpg",
                    "tags": ["classic", "bacon", "simple"],
                },
                {
                    "title": "Fresh Caprese Sandwich",
                    "description": "A simple yet delicious vegetarian sandwich with the flavors of Italy.",
                    "ingredients": "Ciabatta or Focaccia bread, Fresh Mozzarella, Tomatoes, Fresh Basil, Balsamic glaze, Olive oil",
                    "instructions": "1. Slice the bread. 2. Layer mozzarella, tomato slices, and basil leaves. 3. Drizzle with olive oil and balsamic glaze. 4. Serve immediately.",
                    "image_url": "https://www.lekkerensimpel.com/wp-content/uploads/2023/04/588A5318.jpg",
                    "tags": ["vegetarian", "italian", "fresh"],
                },
            ]

    @rx.var
    def selected_recipe(self) -> Optional[Recipe]:
        """Get the selected recipe from the page params."""
        recipe_title = self.router.page.params.get("title", "no-recipe").replace(
            "-", " "
        )
        return next(
            (
                recipe
                for recipe in self.recipes
                if recipe["title"].lower() == recipe_title.lower()
            ),
            None,
        )

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