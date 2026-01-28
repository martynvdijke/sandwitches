import pytest
from django.urls import reverse
from django.utils.translation import activate, gettext as _
from django.utils import translation
from sandwitches.models import User


@pytest.mark.django_db
class TestI18n:
    # --- Existing Tests Cleaned Up ---

    def test_language_selector_present(self, client):
        """Check if the language selection form exists on the settings page."""
        User.objects.create_user(username="testuser", password="password")
        client.login(username="testuser", password="password")
        response = client.get(reverse("user_settings"), follow=True)
        assert response.status_code == 200
        assert b'name="language"' in response.content

    def test_language_selector_not_on_index(self, client):
        """Check if the language selection form no longer exists on the index page."""
        response = client.get(reverse("index"), follow=True)
        assert response.status_code == 200
        assert b'name="language"' not in response.content

    def test_set_language_to_dutch(self, client):
        """Verify that posting to set_language updates the session and cookie."""
        url = reverse("set_language")
        response = client.post(url, {"language": "nl"}, follow=True)

        assert response.status_code == 200
        assert translation.get_language() == "nl"
        assert client.cookies["django_language"].value == "nl"

    def test_set_language_invalid_is_ignored(self, client, settings):
        """Check that unsupported languages (like French) aren't accepted."""
        settings.LANGUAGES = [("en", "English"), ("nl", "Dutch")]
        client.post(reverse("set_language"), {"language": "fr"})

        session_lang = translation.get_language()
        assert session_lang != "fr"

    def test_gettext_function_utility(self):
        """Directly test if gettext finds the translation in the compiled .mo files."""
        activate("nl")
        assert _("Description") == "Beschrijving"

        activate("en")
        assert _("Description") == "Description"
