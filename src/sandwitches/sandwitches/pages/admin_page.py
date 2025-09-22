import reflex as rx
from sandwitches.states.auth_state import AuthState
from sandwitches.components.recipe_upload_form import recipe_upload_form


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
            recipe_upload_form(), class_name="flex justify-center items-start p-8"
        ),
        class_name="bg-gray-50 min-h-screen",
    )