import pytest
from sandwitches.models import Tag


@pytest.mark.django_db
def test_tag_slug_generation():
    t1 = Tag.objects.create(name="Fresh Bread")
    assert t1.slug == "fresh-bread"

    t2 = Tag.objects.create(name="Fresh Bread!")
    # Should handle uniqueness
    assert t2.slug == "fresh-bread-1"

    t3 = Tag.objects.create(name="Fresh Bread?")
    assert t3.slug == "fresh-bread-2"


@pytest.mark.django_db
def test_tag_str():
    t = Tag(name="Cheese")
    assert str(t) == "Cheese"
