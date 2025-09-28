import reflex as rx
from sandwitches.states.auth_state import AuthState
from sandwitches.components.recipe_upload_form import recipe_upload_form
from sandwitches.components.list_recipe import admin_recipe_table
from sandwitches.components.recipe_forms import recipe_edit_form
from sandwitches.states.recipe_state import RecipeState





def admin_page() -> rx.Component:
    return rx.el.div(
        rx.el.header(
            rx.el.div(
                rx.el.h1("Admin Panel", class_name="text-2xl font-bold text-gray-800"),
                rx.el.button(
                    "Sign Out",
                    on_click=AuthState.sign_out,
                    class_name="bg-red-500 text-white px-4 py-2 rounded-lg hover:bg-red-600 transition",
                ),
                class_name="container mx-auto flex justify-between items-center",
            ),
            class_name="bg-white py-4 px-6 border-b border-gray-200",
        ),
        rx.el.div(
            recipe_upload_form(), class_name="flex justify-center items-start p-8",
        ),
        rx.el.div(
            admin_recipe_table(), class_name="flex justify-center items-start p-8",
        ),
        # rx.dialog.root(
        #     rx.dialog.content(
        #         rx.cond(RecipeState.recipe_to_edit, recipe_edit_form(), rx.el.div()),
        #         rx.dialog.close(
        #             rx.el.button(
        #                 rx.icon(tag="x"),
        #                 on_click=RecipeState.cancel_editing,
        #                 class_name="text-gray-500 hover:text-gray-700",
        #                 variant="ghost",
        #             ),
        #             class_name="absolute top-4 right-4",
        #         ),
        #         class_name="bg-white p-8 rounded-2xl shadow-xl border border-gray-100 max-w-2xl w-full",
        #     ),
        #     open=RecipeState.show_edit_modal,
        # ),
        class_name="bg-gray-50 min-h-screen",
    )