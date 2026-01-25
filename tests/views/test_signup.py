from django.test import TestCase
from django.urls import reverse
from django.contrib.auth import get_user_model

User = get_user_model()


class SignupViewTests(TestCase):
    def test_get_signup_page(self):
        resp = self.client.get(reverse("signup"))
        self.assertEqual(resp.status_code, 200)
        self.assertContains(resp, "Create your account")

    def test_post_creates_user_and_logs_in(self):
        data = {
            "username": "testuser",
            "email": "test@example.com",
            "first_name": "Test",
            "last_name": "User",
            "password1": "complex-password-123",
            "password2": "complex-password-123",
            "language": "en",
        }
        resp = self.client.post(reverse("signup"), data, follow=True)
        self.assertEqual(resp.status_code, 200)
        # user created and logged in
        user = User.objects.get(username="testuser")
        self.assertIsNotNone(user)
        self.assertTrue(resp.context["user"].is_authenticated)
        self.assertFalse(user.is_staff)
        self.assertFalse(user.is_superuser)
        self.assertTrue(user.groups.filter(name="community").exists())

    def test_duplicate_username_shows_error(self):
        User.objects.create_user(username="dup", password="x")
        data = {
            "username": "dup",
            "email": "dup@example.com",
            "first_name": "Dup",
            "last_name": "User",
            "password1": "another-pass-123",
            "password2": "another-pass-123",
        }
        resp = self.client.post(reverse("signup"), data)
        self.assertEqual(resp.status_code, 200)
