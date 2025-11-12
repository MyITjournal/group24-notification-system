from django.template import engines

from jinja2 import Template as JinjaTemplate, meta, Environment

def render_jinja(template_str: str, variables: dict):
    env = Environment()
    jinja_template = env.from_string(template_str)
    
    # Find undefined variables
    parsed_content = env.parse(template_str)
    required_vars = meta.find_undeclared_variables(parsed_content)
    missing_vars = required_vars - variables.keys()

    if missing_vars:
        raise ValueError(f"Missing required variables: {', '.join(missing_vars)}")

    return jinja_template.render(**variables)