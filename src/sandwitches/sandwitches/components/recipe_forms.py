import reflex as rx
from sandwitches.states.recipe_state import RecipeState


def form_field(
    label: str, placeholder: str, name: str, field_type: str = "input"
) -> rx.Component:
    return rx.el.div(
        rx.el.label(label, class_name="block text-sm font-medium text-gray-700 mb-1"),
        rx.cond(
            field_type == "textarea",
            rx.el.textarea(
                name=name,
                placeholder=placeholder,
                rows=4,
                class_name="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent transition",
            ),
            rx.el.input(
                name=name,
                placeholder=placeholder,
                class_name="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-orange-500 focus:border-transparent transition",
            ),
        ),
        class_name="mb-4",
    )

def recipe_edit_form() -> rx.Component:
    return rx.el.form(
        rx.el.h2("Edit Recipe", class_name="text-2xl font-bold text-gray-800 mb-6"),
        rx.el.input(
            type="hidden",
            name="id",
            default_value=RecipeState.recipe_to_edit["id"].to(str),
        ),
        form_field(
            "Title",
            "e.g., Classic Club Sandwich",
            "title",
            RecipeState.recipe_to_edit["title"],
        ),
        form_field(
            "Description",
            "A short and sweet description of the recipe.",
            "description",
            RecipeState.recipe_to_edit["description"],
            field_type="textarea",
        ),
        form_field(
            "Tags",
            "e.g., sandwich, classic, lunch",
            "tags",
            RecipeState.recipe_to_edit["tags"].join(", "),
        ),
        form_field(
            "Ingredients",
            "List each ingredient on a new line.",
            "ingredients",
            RecipeState.recipe_to_edit["ingredients"],
            field_type="textarea",
        ),
        form_field(
            "Instructions",
            "Provide step-by-step instructions.",
            "instructions",
            RecipeState.recipe_to_edit["instructions"],
            field_type="textarea",
        ),
        rx.el.div(
            rx.el.label(
                "Recipe Image",
                class_name="block text-sm font-medium text-gray-700 mb-2",
            ),
            rx.upload.root(
                rx.el.div(
                    rx.icon(tag="cloud_upload", class_name="w-10 h-10 text-gray-400"),
                    rx.el.p(
                        "Drag & drop an image here, or click to select",
                        class_name="text-sm text-gray-500",
                    ),
                    class_name="flex flex-col items-center justify-center p-6 border-2 border-dashed border-gray-300 rounded-lg bg-gray-50 text-center cursor-pointer hover:bg-gray-100 transition",
                ),
                id="recipe_image_edit_upload",
                accept={
                    "image/png": [".png"],
                    "image/jpeg": [".jpg", ".jpeg"],
                    "image/webp": [".webp"],
                },
                max_files=1,
                on_drop=RecipeState.handle_edit_upload(
                    rx.upload_files(upload_id="recipe_image_edit_upload")
                ),
                class_name="w-full mb-2",
            ),
            rx.el.div(
                rx.foreach(
                    rx.selected_files("recipe_image_edit_upload"),
                    lambda file: rx.el.div(
                        rx.icon("file-check-2", class_name="text-green-500"),
                        rx.el.span(file),
                        class_name="flex items-center gap-2 p-2 bg-green-50 text-green-700 text-sm rounded-md",
                    ),
                ),
                class_name="space-y-2",
            ),
            class_name="mb-6",
        ),
        rx.el.div(
            rx.el.button(
                "Cancel",
                type="button",
                on_click=RecipeState.cancel_editing,
                class_name="w-full bg-gray-200 text-gray-700 font-bold py-3 rounded-lg hover:bg-gray-300 transition-colors",
            ),
            rx.el.button(
                "Save Changes",
                type="submit",
                class_name="w-full bg-orange-500 text-white font-bold py-3 rounded-lg hover:bg-orange-600 transition-colors",
            ),
            class_name="flex gap-4",
        ),
        on_submit=RecipeState.update_recipe,
        class_name="bg-white p-8 rounded-2xl shadow-xl border border-gray-100 max-w-2xl w-full",
    )

def recipe_upload_form() -> rx.Component:
    return rx.el.form(
        rx.el.h2(
            "Add a New Recipe", class_name="text-2xl font-bold text-gray-800 mb-6"
        ),
        form_field("Title", "e.g., Classic Club Sandwich", "title"),
        form_field(
            "Description",
            "A short and sweet description of the recipe.",
            "description",
            field_type="textarea",
        ),
        form_field("Tags", "e.g., sandwich, classic, lunch", "tags"),
        form_field(
            "Ingredients",
            "List each ingredient on a new line.",
            "ingredients",
            field_type="textarea",
        ),
        form_field(
            "Instructions",
            "Provide step-by-step instructions.",
            "instructions",
            field_type="textarea",
        ),
        rx.el.div(
            rx.el.label(
                "Recipe Image",
                class_name="block text-sm font-medium text-gray-700 mb-2",
            ),
            rx.upload.root(
                rx.el.div(
                    rx.icon(tag="cloud_upload", class_name="w-10 h-10 text-gray-400"),
                    rx.el.p(
                        "Drag & drop an image here, or click to select",
                        class_name="text-sm text-gray-500",
                    ),
                    class_name="flex flex-col items-center justify-center p-6 border-2 border-dashed border-gray-300 rounded-lg bg-gray-50 text-center cursor-pointer hover:bg-gray-100 transition",
                ),
                id="recipe_image_upload",
                accept={
                    "image/png": [".png"],
                    "image/jpeg": [".jpg", ".jpeg"],
                    "image/webp": [".webp"],
                },
                max_files=1,
                multiple=False,
                class_name="w-full mb-2",
            ),
            rx.el.div(
                rx.foreach(
                    rx.selected_files("recipe_image_upload"),
                    lambda file: rx.el.div(
                        rx.icon("file-check-2", class_name="text-green-500"),
                        rx.el.span(file),
                        class_name="flex items-center gap-2 p-2 bg-green-50 text-green-700 text-sm rounded-md",
                    ),
                ),
                class_name="space-y-2",
            ),
            rx.el.button(
                "Upload Image",
                on_click=RecipeState.handle_upload(
                    rx.upload_files(upload_id="recipe_image_upload")
                ),
                type="button",
                class_name="w-full mt-2 bg-gray-200 text-gray-700 font-semibold py-2 rounded-lg hover:bg-gray-300 transition-colors",
            ),
            class_name="mb-6",
        ),
        rx.el.button(
            "Add Recipe",
            type="submit",
            class_name="w-full bg-orange-500 text-white font-bold py-3 rounded-lg hover:bg-orange-600 transition-colors",
        ),
        on_submit=RecipeState.add_recipe,
        reset_on_submit=True,
        class_name="bg-white p-8 rounded-2xl shadow-xl border border-gray-100 max-w-2xl w-full",
    )