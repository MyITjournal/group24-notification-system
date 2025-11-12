# Template Service

A Django-based template rendering service that supports multiple languages, dynamic variables, batch rendering, and optional Kafka integration.

---

## Features

* Create, read, update, and render templates.
* Supports multiple languages and versions.
* Single and batch template rendering.
* Jinja2-based rendering engine.
* Optional Kafka integration for publishing render events.
* Error handling for missing templates, missing variables, or invalid types.
* Health check endpoint with Kafka connectivity status.

---

## Table of Contents

1. [Installation](#installation)
2. [Environment Setup](#environment-setup)
3. [Database](#database)
4. [Running the Service](#running-the-service)
5. [API Endpoints](#api-endpoints)
6. [Testing](#testing)
7. [Deployment](#deployment)

---

## Installation

Clone the repository and create a virtual environment:

```bash
git clone <repository_url>
cd template_service
python -m venv venv
source venv/bin/activate  # Linux/macOS
venv\Scripts\activate     # Windows
```

Install dependencies:

```bash
pip install -r requirements.txt
```

---

## Environment Setup

Create a `.env` file in the root directory:

```env
DEBUG=True
SECRET_KEY=your_secret_key
DJANGO_ALLOWED_HOSTS=localhost 127.0.0.1
DATABASE_URL=postgres://user:password@localhost:5432/template_db
KAFKA_BOOTSTRAP_SERVERS=localhost:9092
```

---

## Database

This project supports **SQLite** for development and **PostgreSQL** for production.

### Migrate Database

```bash
python manage.py makemigrations
python manage.py migrate
```

### Switching between SQLite and PostgreSQL

In `settings.py`, use the `DATABASE_URL` environment variable to configure:

```python
import dj_database_url
DATABASES = {
    "default": dj_database_url.config(default="sqlite:///db.sqlite3")
}
```

---

## Running the Service

### Development server

```bash
python manage.py runserver
```

### Health Check

```http
GET /api/v1/health
```

---

## API Endpoints

1. **Get Template**
   `GET /api/v1/templates/{template_id}?language=en&version=latest`

2. **Create Template**
   `POST /api/v1/templates/`

3. **Render Template**
   `POST /api/v1/templates/{template_id}/render`

4. **Batch Render Templates**
   `POST /api/v1/templates/render/batch`

5. **List Available Templates**
   `GET /api/v1/templates/`

> Full request/response examples are provided in the API documentation.

---

## Testing

Run the tests:

```bash
python manage.py test
```

---

## Deployment

### Using Gunicorn

```bash
gunicorn template_service.wsgi:application --bind 0.0.0.0:8000
```

### With Nginx and SSL

* Configure Nginx as a reverse proxy.
* Set up SSL using Certbot or other providers.

---

## Procfile (for Heroku)

```
web: gunicorn template_service.wsgi:application --log-file -
```

---

## Notes

* Kafka integration is optional; service works without it.
* Batch rendering supports up to 50 templates per request.
* Missing required variables return descriptive errors; preview mode can be used to skip validation.
