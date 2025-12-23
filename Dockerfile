FROM python:3.14

ARG UID=1000
ARG GID=1000
ARG USERNAME=app
ARG GROUPNAME=app

RUN groupadd -g ${GID} ${USERNAME} && \
    useradd -m -u ${UID} -g ${GID} ${GROUPNAME} -s /bin/bash 

RUN mkdir /app
# Set the working directory inside the container
WORKDIR /app

# Set environment variables 
# Prevents Python from writing pyc files to disk
ENV PYTHONDONTWRITEBYTECODE=1
#Prevents Python from buffering stdout and stderr
ENV PYTHONUNBUFFERED=1
ENV PATH="/app/.venv/bin:$PATH"
# Default to non-debug for production images (override at runtime if needed)
ENV DEBUG=0
# Tunables for Gunicorn (can be overridden at runtime)
ENV GUNICORN_WORKERS=3
ENV GUNICORN_THREADS=2

# Ensure runtime server packages are available (installed via pip or project deps)
RUN pip install --upgrade pip && pip install uv gunicorn uvicorn
COPY uv.lock  /app/
COPY pyproject.toml  /app/
COPY . /app/
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
RUN uv sync --locked --no-dev
# Guard: if gunicorn/uvicorn are not in project deps, ensure they're present
RUN pip install --no-deps --upgrade gunicorn uvicorn || true
RUN chown -R root:app /app
ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 6270

USER app

CMD ["/bin/sh", "-c", "python src/manage.py makemigrations sandwitches && python src/manage.py migrate && exec gunicorn sandwitches.asgi:application -k uvicorn.workers.UvicornWorker -b 0.0.0.0:6270 --workers ${GUNICORN_WORKERS} --threads ${GUNICORN_THREADS}"]