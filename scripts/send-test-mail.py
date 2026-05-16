#!/usr/bin/env python3
"""Send a test message to local smtp2http."""
import smtplib
import sys

host = sys.argv[1] if len(sys.argv) > 1 else "127.0.0.1"
port = int(sys.argv[2]) if len(sys.argv) > 2 else 2525
to_addr = sys.argv[3] if len(sys.argv) > 3 else "user@example.com"

msg = (
    "From: sender@test.com\r\n"
    f"To: {to_addr}\r\n"
    "Subject: smtp2http test\r\n"
    "\r\n"
    "Hello from the integration test.\r\n"
)

with smtplib.SMTP(host, port, timeout=30) as smtp:
    smtp.sendmail("sender@test.com", [to_addr], msg)

print(f"Sent test mail to {to_addr} via {host}:{port}")
