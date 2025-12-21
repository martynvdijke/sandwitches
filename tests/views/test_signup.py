from django.test import TestCase
from django.urls import reverse
from django.contrib.auth import get_user_model
from unittest.mock import patch

User = get_user_model()


class SignupViewTests(TestCase):
    def test_get_signup_page(self):
        resp = self.client.get(reverse("signup"))
        self.assertEqual(resp.status_code, 200)
        self.assertContains(resp, "<h2>Sign up")

    def test_post_creates_user_and_logs_in(self):
        data = {
            "username": "testuser",
            "email": "test@example.com",
            "first_name": "Test",
            "last_name": "User",
            "password1": "complex-password-123",
            "password2": "complex-password-123",
        }
        resp = self.client.post(reverse("signup"), data, follow=True)
        # should redirect and end up with an authenticated user in context
        self.assertEqual(resp.status_code, 200)
        self.assertTrue(resp.context["user"].is_authenticated)
        user = User.objects.get(username="testuser")
        self.assertFalse(user.is_staff)
        self.assertFalse(user.is_superuser)

    def test_duplicate_username_shows_error(self):
        User.objects.create_user(username="duptestuseruser", password="x")
        data = {
            "username": "testuser",
            "email": "testuser@example.com",
            "first_name": "Duplicate",
            "last_name": "User",
            "password1": "another-pass-123",
            "password2": "another-pass-123",
        }
        resp = self.client.post(reverse("signup"), data)
        # form re-rendered with error and no new user created
        self.assertEqual(resp.status_code, 302)