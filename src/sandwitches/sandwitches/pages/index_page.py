import reflex as rx
from sandwitches.states.recipe_state import RecipeState
from sandwitches.components.recipe_card import recipe_card


def index_page() -> rx.Component:
    return rx.el.div(
        rx.el.header(
            rx.el.div(
                rx.icon("chef-hat", class_name="h-8 w-8 text-orange-500"),
                rx.el.h1(
                    "Sandwich Sanctuary", class_name="text-3xl font-bold text-gray-800"
                ),
                rx.el.a(
                    rx.el.button(
                        "Admin Login",
                        rx.icon("user-cog", class_name="ml-2"),
                        class_name="bg-orange-500 text-white px-4 py-2 rounded-lg hover:bg-orange-600 transition flex items-center",
                    ),
                    href="/login",
                ),
                class_name="container mx-auto flex justify-between items-center",
            ),
            class_name="bg-white/80 backdrop-blur-md sticky top-0 z-10 py-4 px-4 sm:px-6 lg:px-8 border-b border-gray-200",
        ),
        rx.el.main(
            rx.el.div(
                rx.el.div(
                    rx.el.h2(
                        "Latest Sandwich Creations",
                        class_name="text-2xl font-bold text-gray-800 mb-4",
                    ),
                    rx.el.div(
                        rx.el.span("Filter by tag:", class_name="mr-2 font-medium"),
                        rx.foreach(
                            RecipeState.all_tags,
                            lambda tag: rx.el.button(
                                tag,
                                on_click=lambda: RecipeState.toggle_tag(tag),
                                class_name=rx.cond(
                                    RecipeState.selected_tags.contains(tag),
                                    "bg-orange-500 text-white px-3 py-1 rounded-full text-sm font-semibold transition-colors",
                                    "bg-gray-200 text-gray-700 px-3 py-1 rounded-full text-sm font-semibold hover:bg-gray-300 transition-colors",
                                ),
                                margin_x="1",
                            ),
                        ),
                        class_name="flex items-center mb-8 flex-wrap",
                    ),
                    rx.el.div(
                        rx.foreach(RecipeState.recipes, recipe_card),
                        class_name="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8",
                    ),
                    class_name="container mx-auto",
                ),
                class_name="py-12 px-4 sm:px-6 lg:px-8",
            )
        ),
        class_name="bg-gray-50 min-h-screen",
    )