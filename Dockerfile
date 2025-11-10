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
ENV DEBUG=1

RUN pip install --upgrade pip && pip install uv
COPY uv.lock  /app/
COPY pyproject.toml  /app/
COPY . /app/
RUN uv sync --locked --no-dev
RUN chown -R root:app /app

EXPOSE 6270

USER app

CMD ["/bin/sh", "-c", "python src/sandwitches/manage.py makemigrations sandwitches && python src/sandwitches/manage.py migrate && python src/sandwitches/manage.py runserver 0.0.0.0:6270"]