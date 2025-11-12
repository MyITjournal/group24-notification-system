from django.db import models
import uuid
from django.utils import timezone


class Template(models.Model):
    template_id = models.UUIDField(default=uuid.uuid4, editable=False, unique=True)
    name = models.CharField(max_length=255)
    version = models.CharField(max_length=10, default="latest")
    language = models.CharField(max_length=2, default="en")
    type = models.CharField(max_length=10)
    subject = models.TextField(blank=True, null=True)

    body = models.JSONField(default=dict)  # ✅ Replace body_html + body_text

    variables = models.JSONField(default=list)
    metadata = models.JSONField(default=dict)

    created_at = models.DateTimeField(default=timezone.now)
    updated_at = models.DateTimeField(default=timezone.now)


    class Meta:
        unique_together = ("name", "language", "version")

    def __str__(self):
        return f"{self.name} ({self.language}) v{self.version}"
