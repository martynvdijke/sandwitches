from django.test import TestCase, Client
from django.contrib.auth import get_user_model
from sandwitches.models import Recipe
from django.urls import reverse

User = get_user_model()


class FeatureTests(TestCase):
    def setUp(self):
        self.client = Client()
        self.user1 = User.objects.create_user(username="user1", password="password")
        self.user2 = User.objects.create_user(username="user2", password="password")
        self.superuser = User.objects.create_superuser(
            "admin", "admin@example.com", "strongpassword123"
        )
        self.recipe = Recipe.objects.create(
            title="Test Recipe",
            instructions="Instructions",
            ingredients="Ingredients",
            uploaded_by=self.user1,
        )

        # User1 likes the recipe
        self.user1.favorites.add(self.recipe)

    def test_recipe_liked_by_users_on_index_page(self):
        # Ensure a superuser exists for the index page to not redirect to setup
        response = self.client.get(reverse("index"))
        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "Liked by user1")

    def test_recipe_liked_by_users_on_detail_page(self):
        response = self.client.get(reverse("recipe_detail", args=[self.recipe.slug]))
        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "Liked by user1")

    def test_user_rating_and_comment_on_detail_page(self):
        self.client.login(username="user1", password="password")
        # Submit a rating with a comment
        response = self.client.post(
            reverse("recipe_rate", args=[self.recipe.pk]),
            {
                "score": 8.5,
                "comment": "This is a great recipe!",
            },
        )
        self.assertEqual(response.status_code, 302)  # Should redirect

        # Check if rating and comment are displayed on detail page
        response = self.client.get(reverse("recipe_detail", args=[self.recipe.slug]))
        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "Your rating: <b>8.5</b>")
        self.assertContains(response, "This is a great recipe!")
        self.assertContains(response, "All Ratings")
        self.assertContains(response, "user1")
        self.assertContains(response, "8.5")

    def test_multiple_ratings_and_comments_on_detail_page(self):
        # User1 rates
        self.client.login(username="user1", password="password")
        self.client.post(
            reverse("recipe_rate", args=[self.recipe.pk]),
            {
                "score": 8.0,
                "comment": "Comment by user1",
            },
        )
        self.client.logout()

        # User2 rates
        self.client.login(username="user2", password="password")
        self.client.post(
            reverse("recipe_rate", args=[self.recipe.pk]),
            {
                "score": 9.0,
                "comment": "Comment by user2",
            },
        )
        self.client.logout()

        response = self.client.get(reverse("recipe_detail", args=[self.recipe.slug]))
        self.assertEqual(response.status_code, 200)
        self.assertContains(response, "user1")
        self.assertContains(response, "8.0")
        self.assertContains(response, "Comment by user1")
        self.assertContains(response, "user2")
        self.assertContains(response, "9.0")
        self.assertContains(response, "Comment by user2")

    def test_highlighted_recipes_on_index_page(self):
        # Create a highlighted recipe
        highlighted_recipe = Recipe.objects.create(
            title="Highlighted Sandwich",
            instructions="Make it pop!",
            ingredients="Glitter",
            uploaded_by=self.superuser,
            is_highlighted=True,
        )

        # Ensure a superuser exists (handled in setUp)
        response = self.client.get(reverse("index"))
        self.assertEqual(response.status_code, 200)

        # Check context
        self.assertIn("highlighted_recipes", response.context)
        self.assertIn(highlighted_recipe, response.context["highlighted_recipes"])
        self.assertNotIn(
            self.recipe, response.context["highlighted_recipes"]
        )  # self.recipe is not highlighted

        # Check content
        self.assertContains(response, "Highlighted Sandwich")
