from invoke import task
import os


@task
def linting(c):
    """Run ruff lint checks."""
    print("Running Ruff lint check check...")
    c.run("ruff check src")
    c.run("ruff check tests")
    c.run("ruff format --check src")
    c.run("ruff format --check tests")


@task
def typecheck(c):
    """Run ty type checks."""
    c.run("ty check src")


@task
def formatting(c):
    """Run ruff formatter."""
    print("Running Black formatter...")
    c.run("ruff format src")
    c.run("ruff format tests")
    c.run("ruff check --fix src")
    c.run("ruff check --fix tests")


@task
def tests(c):
    """Run tests with pytest."""
    print("Running tests with pytest...")
    os.environ["SECRET_KEY"] = "tests"
    os.environ["DEBUG"] = "1"
    os.environ["ALLOWED_HOSTS"] = "127.0.0.1"
    os.environ["CSRF_TRUSTED_ORIGINS"] = "http://127.0.0.1"
    c.run("pytest tests")


@task
def ci(c):
    """Run ci checks linting and pytest."""
    linting(c)
    typecheck(c)
    tests(c)
