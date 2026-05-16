#!/usr/bin/env python3
"""Mock webhook: log POST bodies and return HTTP 200."""
from http.server import BaseHTTPRequestHandler, HTTPServer
import sys


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)
        print(f"\n=== POST {self.path} ===", flush=True)
        print(body.decode("utf-8", errors="replace"), flush=True)
        print("=== end ===\n", flush=True)
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"ok")

    def log_message(self, format, *args):
        return


def main():
    host = "127.0.0.1"
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8080
    print(f"Mock webhook listening on http://{host}:{port}/", flush=True)
    HTTPServer((host, port), Handler).serve_forever()


if __name__ == "__main__":
    main()
