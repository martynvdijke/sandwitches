import reflex as rx
from typing import cast


class AuthState(rx.State):
    users: dict[str, str] = {"admin": "password"}
    in_session: bool = False
    error_message: str = ""

    @rx.event
    def sign_in(self, form_data: dict):
        self.error_message = ""
        email = form_data["email"]
        password = form_data["password"]
        if email in self.users and self.users[email] == password:
            self.in_session = True
            return rx.redirect("/admin")
        else:
            self.in_session = False
            self.error_message = "Invalid email or password. Please try again."
            yield rx.toast.error(self.error_message, duration=5000)

    @rx.event
    def sign_out(self):
        self.in_session = False
        return rx.redirect("/")

    def check_session(self):
        return rx.cond(
            self.in_session, rx.noop(), cast(rx.event.EventSpec, rx.redirect("/login"))
        )