
import reflex as rx
from sandwitches.states.recipe_state import RecipeState
from sandwitches.models import Recipe


def admin_recipe_table() -> rx.Component:
    return rx.el.div(
        rx.el.h2("Manage Recipes", class_name="text-2xl font-bold text-gray-800 mb-6"),
        rx.el.div(
            rx.el.table(
                rx.el.thead(
                    rx.el.tr(
                        rx.el.th(
                            "Title",
                            class_name="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider",
                        ),
                        rx.el.th(
                            "Tags",
                            class_name="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider",
                        ),
                        rx.el.th(
                            "Actions",
                            class_name="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider",
                        ),
                    )
                ),
                rx.el.tbody(
                    rx.foreach(RecipeState.recipes, render_recipe_row),
                    class_name="bg-white divide-y divide-gray-200",
                ),
                class_name="min-w-full divide-y divide-gray-200 shadow border-b border-gray-200 sm:rounded-lg",
            ),
            class_name="overflow-x-auto",
        ),
        class_name="bg-white p-8 rounded-2xl shadow-xl border border-gray-100 max-w-4xl w-full",
    )


def render_recipe_row(recipe: Recipe) -> rx.Component:
    return rx.el.tr(
        rx.el.td(
            recipe["title"],
            class_name="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900",
        ),
        rx.el.td(
            rx.el.div(
                rx.foreach(
                    recipe.tags,
                    lambda tag: rx.el.span(
                        tag,
                        class_name="text-xs font-semibold mr-2 px-2.5 py-0.5 rounded-full bg-orange-100 text-orange-800",
                    ),
                ),
                class_name="flex flex-wrap",
            ),
            class_name="px-6 py-4 whitespace-nowrap text-sm text-gray-500",
        ),
        rx.el.td(
            rx.el.a(
                "Edit",
                href="#",
                class_name="text-orange-600 hover:text-orange-900 font-semibold",
            ),
            class_name="px-6 py-4 whitespace-nowrap text-right text-sm font-medium",
        ),
    )