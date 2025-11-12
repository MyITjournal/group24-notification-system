'''from kafka import KafkaProducer
import json
from django.conf import settings
import uuid



producer = KafkaProducer(
    bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
    value_serializer=lambda v: json.dumps(v).encode("utf-8"),
    key_serializer=lambda k: str(k).encode("utf-8") if k else None,
)

def send_render_request(message: dict):
    # Generate correlation ID manually
    message["correlation_id"] = str(uuid.uuid4())

    producer.send(
        settings.KAFKA_TEMPLATE_TOPIC,
        value=message,
        key=message.get("template_id")
    )

    producer.flush()
    return True
'''
import json
import uuid

try:
    from kafka import KafkaProducer
    from django.conf import settings

    producer = KafkaProducer(
        bootstrap_servers=settings.KAFKA_BOOTSTRAP_SERVERS,
        value_serializer=lambda v: json.dumps(v).encode("utf-8"),
        key_serializer=lambda k: str(k).encode("utf-8") if k else None,
    )
    KAFKA_AVAILABLE = True
except Exception:
    # fallback to mock producer
    class MockProducer:
        def send(self, topic, value, key=None):
            print(f"[MOCK-KAFKA] topic={topic}, key={key}")
            print(json.dumps(value, indent=2))

        def flush(self):
            pass

    producer = MockProducer()
    KAFKA_AVAILABLE = False


def send_render_request(message: dict):
    message["correlation_id"] = str(uuid.uuid4())

    topic = (
        settings.KAFKA_TEMPLATE_TOPIC
        if KAFKA_AVAILABLE
        else "mock-topic"
    )

    producer.send(topic, value=message, key=message.get("template_id"))
    producer.flush()

    return True
