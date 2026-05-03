# Stage 1: Build static assets with Node.js
FROM node:24-alpine AS builder

WORKDIR /build
COPY package.json package-lock.json ./
COPY webpack.config.js babel.config.json ./
COPY src/static/ ./src/static/
RUN npm install && npm run build

# Stage 2: Python application
FROM python:3.14-slim

LABEL org.opencontainers.image.source=https://github.com/martynvdijke/sandwitches
LABEL org.opencontainers.image.description="Sandwitches container image"
LABEL org.opencontainers.image.licenses=MIT

ARG UID=1000
ARG GID=1000
ARG USERNAME=app
ARG GROUPNAME=app

RUN groupadd -g ${GID} ${USERNAME} && \
    useradd -m -u ${UID} -g ${GID} ${GROUPNAME} -s /bin/bash && \
    apt-get update && apt-get install -y supervisor --no-install-recommends && \
    apt-get clean && rm -rf /var/lib/apt/lists/* && \
    mkdir /app

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PATH="/app/.venv/bin:$PATH"
ENV GUNICORN_WORKERS=3
ENV GUNICORN_THREADS=2

RUN pip install --upgrade pip --no-cache-dir && pip install uv --no-cache-dir

COPY uv.lock pyproject.toml /app/
RUN uv sync --locked --no-dev --no-cache

COPY . /app/
COPY --from=builder /build/src/static/dist/ /app/src/static/dist/
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh && chown -R root:app /app

ENTRYPOINT ["/app/entrypoint.sh"]

EXPOSE 6270

USER app

CMD ["/bin/sh", "-c", "python src/manage.py collectstatic --noinput --clear && python src/manage.py makemigrations sandwitches && python src/manage.py migrate && supervisord "]
