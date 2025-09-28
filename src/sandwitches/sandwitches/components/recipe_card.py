import reflex as rx
from sandwitches.states.recipe_state import Recipe


def get_image_src(image_url: str) -> rx.Var[str]:
    return rx.cond(
        image_url.startswith("http") | image_url.startswith("/"),
        image_url,
        rx.get_upload_url(image_url),
    )


def recipe_card(recipe: Recipe) -> rx.Component:
    return rx.el.div(
        rx.image(
            src=get_image_src(recipe.image_url),
            class_name="w-full h-48 object-cover",
        ),
        rx.el.div(
            rx.el.div(
                rx.foreach(
                    recipe.tags,
                    lambda tag: rx.el.span(
                        tag,
                        class_name="text-xs font-semibold mr-2 px-2.5 py-0.5 rounded-full bg-orange-100 text-orange-800",
                    ),
                ),
                class_name="flex flex-wrap mb-2",
            ),
            rx.el.h3(recipe.title, class_name="text-lg font-semibold text-gray-800"),
            class_name="p-4",
        ),
        class_name="bg-white rounded-2xl overflow-hidden shadow-sm border border-gray-100 hover:shadow-lg transition-shadow duration-300",
    )