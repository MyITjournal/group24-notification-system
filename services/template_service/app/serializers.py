from rest_framework import serializers
from .models import Template

# ---------- Create/Update serializer ----------
class TemplateCreateSerializer(serializers.ModelSerializer):
    body = serializers.DictField(child=serializers.CharField(), required=True)
    variables = serializers.ListField(child=serializers.DictField(), required=True)
    metadata = serializers.DictField(required=False)

    class Meta:
        model = Template
        fields = [
            "template_id",
            "name",
            "version",
            "language",
            "type",
            "subject",
            "body",
            "variables",
            "metadata"
        ]
        read_only_fields = ["template_id"]

    def validate_variables(self, value):
        for var in value:
            if not all(k in var for k in ("name", "type", "required")):
                raise serializers.ValidationError(
                    "Each variable must have 'name', 'type', and 'required' fields."
                )
        return value


# ---------- Response serializer (exact API shape for GET) ----------
class TemplateResponseSerializer(serializers.Serializer):
    template_id = serializers.CharField()
    name = serializers.CharField()
    version = serializers.CharField()
    language = serializers.CharField()
    type = serializers.CharField()
    subject = serializers.CharField(allow_blank=True, allow_null=True)
    body = serializers.DictField()
    variables = serializers.ListField()
    metadata = serializers.DictField()


# ---------- Render request serializer ----------
class TemplateRenderSerializer(serializers.Serializer):
    language = serializers.CharField(required=False, default="en")
    version = serializers.CharField(required=False, default="latest")
    variables = serializers.DictField(child=serializers.CharField(), required=True)
    preview_mode = serializers.BooleanField(default=False)

    def validate_variables(self, value):
        if not isinstance(value, dict):
            raise serializers.ValidationError("Variables must be a dictionary")
        return value


# ---------- Batch render request ----------
class BatchTemplateRenderRequestSerializer(serializers.Serializer):
    template_id = serializers.CharField()
    language = serializers.CharField(required=False, default="en")
    version = serializers.CharField(required=False, default="latest")
    variables = serializers.DictField(child=serializers.CharField(), required=True)
    preview_mode = serializers.BooleanField(default=False)


class BatchTemplateRenderSerializer(serializers.Serializer):
    requests = BatchTemplateRenderRequestSerializer(many=True)

    def validate_requests(self, value):
        if len(value) > 50:
            raise serializers.ValidationError("Cannot render more than 50 templates at once")
        return value
