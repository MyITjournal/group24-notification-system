
from django.contrib import admin
from django.urls import path,include
from drf_yasg.views import get_schema_view
from drf_yasg import openapi
from rest_framework import permissions

schema_view = get_schema_view(
    openapi.Info(
        title="Template Service API",
        default_version="v1",
        description="Distributed Notification System Template API",
    ),
    public=True,
    permission_classes=(permissions.AllowAny,),
)


urlpatterns = [
    path('admin/', admin.site.urls),
    path("metrics", include("django_prometheus.urls")), #GET /metrics
    path("swagger/", schema_view.with_ui("swagger", cache_timeout=0)),
    path("redoc/", schema_view.with_ui("redoc", cache_timeout=0)),
    path("api/v1/", include("app.urls")),

]
## Endpoints
'''
- POST /api/templates/ — create template
- GET /api/templates/<id>/ — get template
- POST /api/templates/<id>/render/ — render template
- GET /api/templates/health/ — health check
- GET /metrics — prometheus metrics
- GET /swagger/ — swagger docs
- GET /redoc/ — redoc docs'''