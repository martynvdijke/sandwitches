FROM python:3.14

RUN mkdir /app
# Set the working directory inside the container
WORKDIR /app

# Set environment variables 
# Prevents Python from writing pyc files to disk
ENV PYTHONDONTWRITEBYTECODE=1
#Prevents Python from buffering stdout and stderr
ENV PYTHONUNBUFFERED=1
ENV PATH="/app/.venv/bin:$PATH"
ENV DEBUG=1

RUN pip install --upgrade pip && pip install uv

COPY uv.lock  /app/
COPY pyproject.toml  /app/
COPY . /app/

RUN uv sync --locked --no-dev

EXPOSE 6270

CMD ["/bin/sh", "-c", "python src/sandwitches/manage.py makemigrations sandwitches && python src/sandwitches/manage.py migrate && python src/sandwitches/manage.py runserver 0.0.0.0:8000"]