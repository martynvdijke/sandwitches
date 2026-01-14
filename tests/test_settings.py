from django.test import TestCase
from sandwitches.models import Setting
from django.contrib.auth import get_user_model
from ninja.testing import TestClient
from ninja import NinjaAPI
from sandwitches.api import api
import json

User = get_user_model()

class SettingsTest(TestCase):
    def setUp(self):
        NinjaAPI._registry = []
        self.client = TestClient(api)
        Setting.objects.get_or_create()
        self.user = User.objects.create_user(username="user", password="password")
        self.staff = User.objects.create_user(
            username="staff", password="password", is_staff=True
        )

    def test_singleton_instance(self):
        s1 = Setting.objects.get()
        s2 = Setting()
        s2.site_name = "Another Name"
        s2.save()
        self.assertEqual(s1.pk, s2.pk)
        self.assertEqual(Setting.objects.count(), 1)

    def test_get_settings_unauthenticated(self):
        response = self.client.get("/v1/settings")
        self.assertEqual(response.status_code, 200)

    def test_update_settings_unauthenticated(self):
        response = self.client.post("/v1/settings", json={})
        self.assertEqual(response.status_code, 401)

    def test_update_settings_as_non_staff(self):
        response = self.client.post(
            "/v1/settings",
            json={"site_name": "New Name"},
            auth=self.user,
        )
        self.assertEqual(response.status_code, 403)

    def test_update_settings_as_staff(self):
        response = self.client.post(
            "/v1/settings",
            json={"site_name": "New Name"},
            auth=self.staff,
        )
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json()["site_name"], "New Name")
        self.assertEqual(Setting.objects.get().site_name, "New Name")