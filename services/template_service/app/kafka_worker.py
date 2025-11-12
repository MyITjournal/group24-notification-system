import json
from kafka import KafkaConsumer, KafkaProducer
from django.conf import settings
import django
import os
os.environ.setdefault("DJANGO_SETTINGS_MODULE", "app.settings")
django.setup()

from .jinja_renderer import render_jinja
from app.models import Template

consumer = KafkaConsumer(
    settings.KAFKA_TEMPLATE_TOPIC,
    bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
    value_deserializer=lambda v: json.loads(v.decode("utf-8")),
)

producer = KafkaProducer(
    bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
)

for msg in consumer:
    data = msg.value
    template_id = data["template_id"]
    variables = data.get("variables", {})

    try:
        template = Template.objects.get(template_id=template_id)
        rendered = render_jinja(template.content, variables)

        output = {
            "template_id": template_id,
            "rendered": rendered,
            "correlation_id": data.get("correlation_id")
        }

        producer.send(settings.KAFKA_RESPONSE_TOPIC, value=output)
        producer.flush()

    except Template.DoesNotExist:
        producer.send(settings.KAFKA_RESPONSE_TOPIC, value={
            "error": "template_not_found",
            "template_id": template_id,
            "correlation_id": data.get("correlation_id")
        })
