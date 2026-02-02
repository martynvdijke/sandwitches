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

    # Check for error messages if any
    messages = list(resp.context["messages"]) if resp.context else []
    for m in messages:
        print(f"Message: {m}")

    assert not CartItem.objects.filter(user=user).exists()

    # Should have created 1 order
    assert Order.objects.filter(user=user).exists()
    order = Order.objects.filter(user=user).first()
    assert order.items.count() == 1
    item = order.items.first()
    assert item.recipe == recipe
    assert item.quantity == 3
    assert order.total_price == 30.00


@pytest.mark.django_db
def test_cart_multiple_items_checkout(client, user_factory):
    user = user_factory(username="multi_cart_user")
    client.force_login(user)

    r1 = Recipe.objects.create(title="S1", price=10.00, is_approved=True)
    r2 = Recipe.objects.create(title="S2", price=20.00, is_approved=True)

    # Add both to cart
    client.post(reverse("add_to_cart", kwargs={"pk": r1.pk}))
    client.post(reverse("add_to_cart", kwargs={"pk": r2.pk}))

    assert CartItem.objects.filter(user=user).count() == 2

    # Checkout
    resp = client.post(reverse("checkout_cart"), follow=True)
    assert resp.status_code == 200

    # Check messages
    messages = (
        [m.message for m in list(resp.context["messages"])] if resp.context else []
    )
    assert "Orders submitted successfully!" in messages

    assert not CartItem.objects.filter(user=user).exists()
    assert Order.objects.filter(user=user).count() == 1
    order = Order.objects.get(user=user)
    assert order.items.count() == 2
    assert order.total_price == 30.00


@pytest.mark.django_db
def test_add_non_orderable_recipe_fails(client, user_factory):
    user = user_factory()
    recipe = Recipe.objects.create(title="Free Sandwich", price=None)

    client.force_login(user)
    url_add = reverse("add_to_cart", kwargs={"pk": recipe.pk})
    resp = client.post(url_add, follow=True)
    assert not CartItem.objects.filter(user=user, recipe=recipe).exists()
    assert "cannot be ordered" in resp.content.decode()
