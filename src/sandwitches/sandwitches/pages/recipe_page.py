import reflex as rx
from sandwitches.states.recipe_state import RecipeState
from sandwitches.components.recipe_card import get_image_src


def recipe_detail_view() -> rx.Component:
    return rx.el.div(
        rx.cond(
            RecipeState.selected_recipe,
            rx.el.div(
                rx.el.div(
                    rx.el.a(
                        rx.icon(tag="arrow-left", class_name="mr-2"),
                        "Back to Recipes",
                        href="/",
                        class_name="flex items-center text-orange-500 hover:text-orange-600 font-semibold mb-8",
                    ),
                    rx.el.h1(
                        RecipeState.selected_recipe["title"],
                        class_name="text-4xl font-bold text-gray-800 mb-4",
                    ),
                    rx.el.div(
                        rx.foreach(
                            RecipeState.selected_recipe["tags"],
                            lambda tag: rx.el.span(
                                tag,
                                class_name="text-sm font-semibold mr-2 px-3 py-1 rounded-full bg-orange-100 text-orange-800",
                            ),
                        ),
                        class_name="flex flex-wrap mb-6",
                    ),
                    rx.image(
                        src=get_image_src(RecipeState.selected_recipe["image_url"]),
                        class_name="w-full h-96 object-cover rounded-2xl mb-8 shadow-lg",
                    ),
                    rx.el.p(
                        RecipeState.selected_recipe["description"],
                        class_name="text-lg text-gray-700 mb-8",
                    ),
                    rx.el.div(
                        rx.el.h2(
                            "Ingredients",
                            class_name="text-2xl font-bold text-gray-800 mb-4",
                        ),
                        rx.el.ul(
                            rx.foreach(
                                RecipeState.selected_recipe["ingredients"].split("\n"),
                                lambda ingredient: rx.el.li(
                                    ingredient,
                                    class_name="text-gray-700 mb-2 pl-4 border-l-2 border-orange-500",
                                ),
                            ),
                            class_name="list-none mb-8",
                        ),
                        class_name="bg-gray-50 p-6 rounded-lg",
                    ),
                    rx.el.div(
                        rx.el.h2(
                            "Instructions",
                            class_name="text-2xl font-bold text-gray-800 my-4",
                        ),
                        rx.el.div(
                            rx.foreach(
                                RecipeState.selected_recipe["instructions"].split("\n"),
                                lambda instruction, index: rx.el.div(
                                    rx.el.span(
                                        f"{index + 1}.",
                                        class_name="font-bold text-orange-500 mr-4",
                                    ),
                                    rx.el.p(
                                        instruction, class_name="text-gray-700 flex-1"
                                    ),
                                    class_name="flex items-start mb-4",
                                ),
                            )
                        ),
                        class_name="mt-8",
                    ),
                    class_name="max-w-4xl mx-auto",
                )
            ),
            rx.el.div(
                rx.el.h2(
                    "Recipe not found", class_name="text-2xl font-bold text-gray-800"
                ),
                rx.el.p(
                    "The recipe you are looking for does not exist or could not be loaded.",
                    class_name="mt-4",
                ),
                rx.el.a(
                    "Go back to recipes",
                    href="/",
                    class_name="text-orange-500 mt-4 inline-block",
                ),
                class_name="text-center py-20",
            ),
        ),
        class_name="py-12 px-4 sm:px-6 lg:px-8",
    )


def recipe_page() -> rx.Component:
    return rx.el.div(
        rx.el.header(
            rx.el.div(
                rx.el.a(
                    rx.icon("chef-hat", class_name="h-8 w-8 text-orange-500"), href="/"
                ),
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
        rx.el.main(recipe_detail_view()),
        class_name="bg-gray-50 min-h-screen font-['Inter']",
    )