import json
from django.shortcuts import get_object_or_404
from rest_framework.views import APIView
from rest_framework.response import Response
from rest_framework import status
from django.conf import settings
from .models import Template
from .serializers import (
    TemplateCreateSerializer,
    TemplateResponseSerializer,
    TemplateRenderSerializer,
    BatchTemplateRenderSerializer,
)
from .jinja_renderer import render_jinja  # keep your renderer; used for actual rendering
from .kafka_producer import send_render_request  # optional; wrapped safely
from dateutil.parser import parse as parse_date
from typing import Tuple, List, Dict


# ---------- Helpers ----------
def get_request_id(request):
    return request.headers.get("X-Request-ID") or request.query_params.get("request_id")


def format_template_response(obj: Template) -> Dict:
    """
    Produce the exact documentation response shape for GET /templates/{id}.
    Fill missing metadata fields with sensible defaults when absent.
    """
    body = {
        "html": obj.body.get("html") if isinstance(obj.body, dict) else None,
        "text": obj.body.get("text") if isinstance(obj.body, dict) else obj.body or ""
    } if getattr(obj, "body", None) is not None else {"html": None, "text": ""}

    variables_out = []
    for v in (obj.variables or []):
        variables_out.append({
            "name": v.get("name"),
            "type": v.get("type"),
            "required": v.get("required", False),
            "description": v.get("description", "")  # default empty description
        })

    metadata = obj.metadata.copy() if isinstance(obj.metadata, dict) else {}
    # Ensure timestamps exist in metadata (use model timestamps)
    metadata.setdefault("created_at", obj.created_at.isoformat() if hasattr(obj, "created_at") else None)
    metadata.setdefault("updated_at", obj.updated_at.isoformat() if hasattr(obj, "updated_at") else None)
    metadata.setdefault("created_by", metadata.get("created_by", None))
    metadata.setdefault("tags", metadata.get("tags", []))

    return {
        "template_id": str(obj.template_id) if getattr(obj, "template_id", None) else obj.name,
        "name": obj.name,
        "version": obj.version,
        "language": obj.language,
        "type": obj.type,
        "subject": obj.subject,
        "body": body,
        "variables": variables_out,
        "metadata": metadata
    }


def type_check(variable_def: Dict, value) -> Tuple[bool, Dict]:
    """
    Return (ok: bool, error_detail: dict)
    variable_def: {"name":..., "type": "string"|"number"|"boolean"|"date", ...}
    """
    expected = (variable_def.get("type") or "string").lower()
    name = variable_def.get("name")
    if value is None:
        return True, {}
    # string
    if expected == "string":
        if isinstance(value, str):
            return True, {}
        return False, {"variable": name, "expected_type": "string", "received_type": type(value).__name__, "received_value": str(value)}
    if expected == "number":
        if isinstance(value, (int, float)):
            return True, {}
        # try to coerce numeric string
        if isinstance(value, str):
            try:
                float(value)
                return True, {}
            except:
                return False, {"variable": name, "expected_type": "number", "received_type": "string", "received_value": value}
        return False, {"variable": name, "expected_type": "number", "received_type": type(value).__name__, "received_value": str(value)}
    if expected == "boolean":
        if isinstance(value, bool):
            return True, {}
        if isinstance(value, str) and value.lower() in ("true", "false", "0", "1"):
            return True, {}
        return False, {"variable": name, "expected_type": "boolean", "received_type": type(value).__name__, "received_value": str(value)}
    if expected == "date":
        if isinstance(value, str):
            try:
                parse_date(value)
                return True, {}
            except:
                return False, {"variable": name, "expected_type": "date", "received_type": "string", "received_value": value}
        return False, {"variable": name, "expected_type": "date", "received_type": type(value).__name__, "received_value": str(value)}
    # default: accept
    return True, {}


def prepare_variables_with_preview(template_obj: Template, variables: dict, preview_mode: bool) -> Tuple[dict, List[str]]:
    """
    Returns variables dict used for rendering and list of missing required var names.
    If preview_mode is True, missing required vars are replaced with placeholders like "{{ var_name }}".
    """
    vars_out = dict(variables or {})
    missing = []
    for v in (template_obj.variables or []):
        name = v.get("name")
        required = v.get("required", False)
        if name not in vars_out or vars_out.get(name) in (None, ""):
            if required:
                missing.append(name)
                if preview_mode:
                    vars_out[name] = "{{ " + name + " }}"
            else:
                # optional: leave as-is or set to empty
                vars_out.setdefault(name, "")
    return vars_out, missing


# ---------- Views ----------

class HealthCheckView(APIView):
    authentication_classes = []
    permission_classes = []

    def get(self, request):
        kafka_status = "unreachable"
        try:
            from kafka import KafkaAdminClient
            client = KafkaAdminClient(
                bootstrap_servers=getattr(settings, "KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
                client_id="health_check"
            )
            client.close()
            kafka_status = "connected"
        except Exception:
            kafka_status = "unreachable"

        return Response({
            "status": "healthy",
            "service": "template_service",
            "kafka": kafka_status
        })


class TemplateCreateView(APIView):
    def post(self, request):
        serializer = TemplateCreateSerializer(data=request.data)
        if not serializer.is_valid():
            return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)
        obj = serializer.save()
        resp = format_template_response(obj)
        return Response(resp, status=status.HTTP_201_CREATED)


class TemplateDetailView(APIView):
    def get(self, request, template_id):
        # support template_id as string name or UUID
        try:
            template = Template.objects.get(template_id=template_id)
        except Template.DoesNotExist:
            # try by name fallback (some specs use code string)
            try:
                template = Template.objects.get(name=template_id)
            except Template.DoesNotExist:
                request_id = get_request_id(request)
                return Response({
                    "error": {
                        "code": "TEMPLATE_NOT_FOUND",
                        "message": f"Template '{template_id}' does not exist",
                        "details": {
                            "template_id": template_id,
                            "available_templates": list(Template.objects.values_list("name", flat=True)[:10])
                        },
                        "request_id": request_id
                    }
                }, status=status.HTTP_404_NOT_FOUND)

        # language validation
        language = request.query_params.get("language", "en")
        supported = getattr(template, "languages", None) or []
        # if template has languages field (list) stored, check it
        if supported and language not in supported:
            request_id = get_request_id(request)
            return Response({
                "error": {
                    "code": "INVALID_LANGUAGE",
                    "message": f"Language '{language}' is not supported",
                    "details": {
                        "requested_language": language,
                        "supported_languages": supported
                    },
                    "request_id": request_id
                }
            }, status=status.HTTP_422_UNPROCESSABLE_ENTITY)

        resp = format_template_response(template)
        return Response(resp, status=status.HTTP_200_OK)


class TemplateRenderView(APIView):
    def post(self, request, template_id):
        template = None
        try:
            template = Template.objects.get(template_id=template_id)
        except Template.DoesNotExist:
            try:
                template = Template.objects.get(name=template_id)
            except Template.DoesNotExist:
                request_id = get_request_id(request)
                return Response({
                    "error": {
                        "code": "TEMPLATE_NOT_FOUND",
                        "message": f"Template '{template_id}' does not exist",
                        "details": {"template_id": template_id},
                        "request_id": request_id
                    }
                }, status=status.HTTP_404_NOT_FOUND)

        serializer = TemplateRenderSerializer(data=request.data)
        if not serializer.is_valid():
            return Response(serializer.errors, status=status.HTTP_400_BAD_REQUEST)

        data = serializer.validated_data
        preview_mode = data.get("preview_mode", False)
        variables_in = data.get("variables", {})

        # prepare variables and detect missing required ones
        vars_for_render, missing = prepare_variables_with_preview(template, variables_in, preview_mode)

        if missing and not preview_mode:
            request_id = get_request_id(request)
            return Response({
                "error": {
                    "code": "MISSING_VARIABLES",
                    "message": "Required variables are missing",
                    "details": {
                        "missing_variables": missing,
                        "provided_variables": list(variables_in.keys())
                    },
                    "request_id": request_id
                }
            }, status=status.HTTP_422_UNPROCESSABLE_ENTITY)

        # type checking
        for vdef in (template.variables or []):
            if vdef.get("name") in vars_for_render:
                ok, err = type_check(vdef, vars_for_render.get(vdef.get("name")))
                if not ok:
                    request_id = get_request_id(request)
                    return Response({
                        "error": {
                            "code": "INVALID_VARIABLE_TYPE",
                            "message": f"Variable '{err.get('variable')}' has invalid type",
                            "details": err,
                            "request_id": request_id
                        }
                    }, status=status.HTTP_422_UNPROCESSABLE_ENTITY)

        # render subject and bodies
        try:
            subject_template = template.subject or ""
            body_html_template = template.body.get("html") if isinstance(template.body, dict) else template.body or ""
            body_text_template = template.body.get("text") if isinstance(template.body, dict) else ""

            rendered_subject = render_jinja(subject_template, vars_for_render)
            rendered_html = render_jinja(body_html_template, vars_for_render)
            rendered_text = render_jinja(body_text_template, vars_for_render)

            response_payload = {
                "template_id": str(template.template_id) if getattr(template, "template_id", None) else template.name,
                "language": data.get("language", template.language),
                "version": data.get("version", template.version),
                "rendered": {
                    "subject": rendered_subject,
                    "body": {
                        "html": rendered_html,
                        "text": rendered_text
                    }
                },
                "rendered_at": template.updated_at.isoformat() if hasattr(template, "updated_at") else None,
                "variables_used": list(vars_for_render.keys())
            }

            # non-blocking Kafka publish
            try:
                send_render_request({
                    "template_id": response_payload["template_id"],
                    "language": response_payload["language"],
                    "version": response_payload["version"],
                    "rendered": response_payload["rendered"],
                    "request_id": get_request_id(request)
                })
            except Exception:
                # swallow kafka errors in rendering flow
                pass

            return Response(response_payload, status=status.HTTP_200_OK)

        except Exception as e:
            request_id = get_request_id(request)
            return Response({
                "error": {
                    "code": "RENDER_ERROR",
                    "message": str(e),
                    "request_id": request_id
                }
            }, status=status.HTTP_500_INTERNAL_SERVER_ERROR)


class TemplateBatchRenderView(APIView):
    def post(self, request):
        batch_serializer = BatchTemplateRenderSerializer(data=request.data)
        if not batch_serializer.is_valid():
            return Response({"error": batch_serializer.errors}, status=status.HTTP_400_BAD_REQUEST)

        results = []
        total_success = 0
        total_failed = 0

        for item in batch_serializer.validated_data["requests"]:
            template_id = item.get("template_id")
            language = item.get("language", "en")
            version = item.get("version", "latest")
            variables = item.get("variables", {})
            preview_mode = item.get("preview_mode", False)

            # try to find template (by UUID or name)
            try:
                template = Template.objects.get(template_id=template_id)
            except Template.DoesNotExist:
                try:
                    template = Template.objects.get(name=template_id)
                except Template.DoesNotExist:
                    results.append({
                        "template_id": template_id,
                        "success": False,
                        "error": {
                            "code": "TEMPLATE_NOT_FOUND",
                            "message": f"Template '{template_id}' does not exist"
                        }
                    })
                    total_failed += 1
                    continue

            vars_for_render, missing = prepare_variables_with_preview(template, variables, preview_mode)
            if missing and not preview_mode:
                results.append({
                    "template_id": template_id,
                    "success": False,
                    "error": {
                        "code": "MISSING_VARIABLES",
                        "message": "Required variables are missing",
                        "details": {
                            "missing_variables": missing,
                            "provided_variables": list(variables.keys())
                        }
                    }
                })
                total_failed += 1
                continue

            # type checking
            bad_type = None
            bad_error = None
            for vdef in (template.variables or []):
                if vdef.get("name") in vars_for_render:
                    ok, err = type_check(vdef, vars_for_render.get(vdef.get("name")))
                    if not ok:
                        bad_type = True
                        bad_error = err
                        break
            if bad_type:
                results.append({
                    "template_id": template_id,
                    "success": False,
                    "error": {
                        "code": "INVALID_VARIABLE_TYPE",
                        "message": f"Variable '{bad_error.get('variable')}' has invalid type",
                        "details": bad_error
                    }
                })
                total_failed += 1
                continue

            # render
            try:
                subject_template = template.subject or ""
                body_html_template = template.body.get("html") if isinstance(template.body, dict) else template.body or ""
                body_text_template = template.body.get("text") if isinstance(template.body, dict) else ""

                rendered_subject = render_jinja(subject_template, vars_for_render)
                rendered_html = render_jinja(body_html_template, vars_for_render)
                rendered_text = render_jinja(body_text_template, vars_for_render)

                results.append({
                    "template_id": template_id,
                    "success": True,
                    "rendered": {
                        "subject": rendered_subject,
                        "body": {
                            "html": rendered_html,
                            "text": rendered_text
                        }
                    }
                })
                total_success += 1
            except Exception as e:
                results.append({
                    "template_id": template_id,
                    "success": False,
                    "error": {
                        "code": "RENDER_ERROR",
                        "message": str(e)
                    }
                })
                total_failed += 1

        return Response({
            "results": results,
            "total_requested": len(batch_serializer.validated_data["requests"]),
            "total_success": total_success,
            "total_failed": total_failed
        }, status=status.HTTP_200_OK)
