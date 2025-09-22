import reflex as rx
from sandwitches.states.auth_state import AuthState


def login_form() -> rx.Component:
    return rx.el.div(
        rx.el.div(
            rx.icon("chef-hat", class_name="h-10 w-10 text-orange-500"),
            rx.el.h2("Admin Login", class_name="text-2xl font-bold text-gray-800"),
            class_name="flex flex-col items-center gap-4 mb-6",
        ),
        rx.el.form(
            rx.el.div(
                rx.el.label("Username", class_name="text-sm font-medium text-gray-600"),
                rx.el.input(
                    placeholder="admin",
                    id="email",
                    type="text",
                    class_name="w-full px-4 py-2 mt-1 rounded-lg border border-gray-300 focus:ring-2 focus:ring-orange-400 focus:border-transparent transition",
                ),
                class_name="mb-4",
            ),
            rx.el.div(
                rx.el.label("Password", class_name="text-sm font-medium text-gray-600"),
                rx.el.input(
                    placeholder="password",
                    id="password",
                    type="password",
                    class_name="w-full px-4 py-2 mt-1 rounded-lg border border-gray-300 focus:ring-2 focus:ring-orange-400 focus:border-transparent transition",
                ),
                class_name="mb-6",
            ),
            rx.el.button(
                "Sign In",
                rx.icon("log-in", class_name="ml-2"),
                type="submit",
                class_name="w-full bg-orange-500 text-white font-semibold py-3 rounded-lg hover:bg-orange-600 transition-colors duration-300 flex items-center justify-center",
            ),
            on_submit=AuthState.sign_in,
        ),
        class_name="bg-white p-8 rounded-2xl shadow-lg border border-gray-100 w-full max-w-sm",
    )