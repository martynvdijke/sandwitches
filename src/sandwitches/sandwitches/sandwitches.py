import reflex as rx
from sandwitches.pages.index_page import index_page
from sandwitches.pages.login_page import login_page
from sandwitches.pages.admin_page import admin_page
from sandwitches.states.auth_state import AuthState
from sandwitches.states.recipe_state import RecipeState


def index() -> rx.Component:
    return index_page()


def login() -> rx.Component:
    return login_page()


def admin() -> rx.Component:
    return admin_page()


app = rx.App(
    theme=rx.theme(appearance="light"),
    head_components=[
        rx.el.link(rel="preconnect", href="https://fonts.googleapis.com"),
        rx.el.link(rel="preconnect", href="https://fonts.gstatic.com", cross_origin=""),
        rx.el.link(
            href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap",
            rel="stylesheet",
        ),
    ],
)
app.add_page(index, route="/", on_load=RecipeState.load_recipes)
app.add_page(login, route="/login")
app.add_page(admin, route="/admin", on_load=AuthState.check_session)