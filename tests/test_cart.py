import pytest
from django.urls import reverse
from sandwitches.models import Recipe, CartItem, Order, User
from django.contrib.auth.models import Group


@pytest.mark.django_db
def test_cart_functionality(client, user_factory):
    # Setup user and groups
    user = user_factory(username="cart_user")
    community_group, _ = Group.objects.get_or_create(name="community")
    user.groups.add(community_group)

    admin_user = User.objects.create_superuser("admin_cart", "admin@cart.com", "pw")
    admin_group, _ = Group.objects.get_or_create(name="admin")
    admin_user.groups.add(admin_group)

    # Setup recipe
    recipe = Recipe.objects.create(
        title="Orderable Sandwich",
        price=10.00,
        uploaded_by=admin_user,
        is_approved=True,
    )

    # 1. Login
    client.force_login(user)

    # 2. Add to cart
    url_add = reverse("add_to_cart", kwargs={"pk": recipe.pk})
    resp = client.post(url_add, follow=True)
    assert resp.status_code == 200
    assert CartItem.objects.filter(user=user, recipe=recipe).exists()
    assert CartItem.objects.get(user=user, recipe=recipe).quantity == 1

    # 3. Add again (increase quantity)
    client.post(url_add, follow=True)
    assert CartItem.objects.get(user=user, recipe=recipe).quantity == 2

    # 4. View cart
    url_cart = reverse("view_cart")
    resp = client.get(url_cart)
    assert resp.status_code == 200
    assert "Orderable Sandwich" in resp.content.decode()
    assert "20.0" in resp.content.decode()  # total_price = 10 * 2

    # 5. Update quantity
    cart_item = CartItem.objects.get(user=user, recipe=recipe)
    url_update = reverse("update_cart_quantity", kwargs={"pk": cart_item.pk})
    client.post(url_update, {"quantity": 3})
    cart_item.refresh_from_db()
    assert cart_item.quantity == 3

    # 6. Checkout
    url_checkout = reverse("checkout_cart")
    resp = client.post(url_checkout, follow=True)
    assert resp.status_code == 200
    assert not CartItem.objects.filter(user=user).exists()
    # Should have created 3 orders
    assert Order.objects.filter(user=user, recipe=recipe).count() == 3


@pytest.mark.django_db
def test_add_non_orderable_recipe_fails(client, user_factory):
    user = user_factory()
    recipe = Recipe.objects.create(title="Free Sandwich", price=None)

    client.force_login(user)
    url_add = reverse("add_to_cart", kwargs={"pk": recipe.pk})
    resp = client.post(url_add, follow=True)
    assert not CartItem.objects.filter(user=user, recipe=recipe).exists()
    assert "cannot be ordered" in resp.content.decode()
