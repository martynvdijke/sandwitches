from invoke import task


@task
def linting(c):
    """Run ruff lint checks."""
    print("Running Ruff lint check check...")
    c.run("ruff check src")
    c.run("ruff check tests")
    c.run("ruff format --check src")
    c.run("ruff format --check tests")

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
    c.run("pytest tests")

@task
def ci(c):
    """Run ci checks linting and pytest."""
    linting(c)
    tests(c)
