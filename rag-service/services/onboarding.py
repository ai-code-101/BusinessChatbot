"""
Validation helpers for onboarding form submissions.
"""

import re

PHONE_RE = re.compile(r"^\+?[0-9\s\-()]{7,15}$")


def validate_details(name: str, phone: str) -> tuple[bool, str | None]:
    if not name or not name.strip():
        return False, "Name is required."
    if not phone or not PHONE_RE.match(phone.strip()):
        return False, "Please enter a valid phone number."
    return True, None