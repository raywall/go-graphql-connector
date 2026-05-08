from http.server import BaseHTTPRequestHandler, HTTPServer
import json


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/health":
            self.respond(200, {"status": "ok"})
            return

        if self.path.startswith("/limites/"):
            codigo = self.path.rsplit("/", 1)[-1]
            if codigo == "10341":
                self.respond(
                    200,
                    {
                        "percentualMaximoMargemConsignavel": 0.35,
                        "origem": "api",
                    },
                )
                return

        self.respond(404, {"message": "not found"})

    def log_message(self, format, *args):
        return

    def respond(self, status, payload):
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


HTTPServer(("0.0.0.0", 8090), Handler).serve_forever()
