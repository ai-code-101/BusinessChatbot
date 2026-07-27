"""
Sends onboarding notification emails via Gmail SMTP.
"""

import os
import smtplib
from email.mime.text import MIMEText

SMTP_HOST = "smtp.gmail.com"
SMTP_PORT = 587

SENDER_EMAIL = "roadyridge@gmail.com"
RECIPIENT_EMAIL = "Sales@peakmobile.co.ke"


def send_onboarding_email(name: str, phone: str, business_id: str = "") -> None:
    sender_password = os.getenv("GMAIL_APP_PASSWORD")
    if not sender_password:
        raise RuntimeError("GMAIL_APP_PASSWORD environment variable is not set")

    body_lines = [
        f"Name: {name}",
        f"Phone: {phone}",
    ]
    if business_id:
        body_lines.append(f"Business ID: {business_id}")

    msg = MIMEText("\n".join(body_lines))
    msg["Subject"] = "customer onboarding"
    msg["From"] = SENDER_EMAIL
    msg["To"] = RECIPIENT_EMAIL

    with smtplib.SMTP(SMTP_HOST, SMTP_PORT, timeout=15) as server:
        server.starttls()
        server.login(SENDER_EMAIL, sender_password)
        server.sendmail(SENDER_EMAIL, [RECIPIENT_EMAIL], msg.as_string())