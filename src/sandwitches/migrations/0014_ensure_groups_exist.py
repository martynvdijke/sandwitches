from django.db import migrations


def create_groups(apps, schema_editor):
    Group = apps.get_model("auth", "Group")
    Group.objects.get_or_create(name="admin")
    Group.objects.get_or_create(name="community")


def remove_groups(apps, schema_editor):
    Group = apps.get_model("auth", "Group")
    Group.objects.filter(name__in=["admin", "community"]).delete()


class Migration(migrations.Migration):
    dependencies = [
        ("sandwitches", "0013_cartitem"),
    ]

    operations = [
        migrations.RunPython(create_groups, remove_groups),
    ]
