from django.urls import path
from .views import (
    HealthCheckView,
    TemplateDetailView,
    TemplateRenderView,
    TemplateBatchRenderView,
    TemplateListCreateView
)

urlpatterns = [
    # Health check
    path("health/", HealthCheckView.as_view(), name="health-check"),

    # Create a template
    path("templates/", TemplateListCreateView.as_view(), name="template-list-create"),
    # Get template by ID or name
    path("templates/<str:template_id>/", TemplateDetailView.as_view(), name="template-detail"),

    # Render single template
    path("templates/<str:template_id>/render/", TemplateRenderView.as_view(), name="template-render"),

    # Batch render
    path("templates/render/batch/", TemplateBatchRenderView.as_view(), name="template-batch-render"),
]
