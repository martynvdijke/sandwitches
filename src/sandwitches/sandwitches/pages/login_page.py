import reflex as rx
from sandwitches.components.login_form import login_form


def login_page() -> rx.Component:
    return rx.el.div(
        login_form(),
        class_name="min-h-screen flex items-center justify-center bg-gray-50 p-4",
    )