#! /bin/python3
"""
WSGI Stdin/Stdout Runner

This script loads a WSGI application and processes HTTP-like requests via stdin/stdout
rather than a network socket. It's used as a bridge between interchange and Python WSGI apps
"""

import sys
import io
import importlib
from typing import Dict, Tuple, Iterable


def load_wsgi_app(module_path: str):
    mod = importlib.import_module(module_path)
    if hasattr(mod, 'application'):
        return getattr(mod, 'application')
    if hasattr(mod, 'app'):
        return getattr(mod, 'app')
    if hasattr(mod, 'create_app') and callable(mod.create_app):
        return mod.create_app()
    raise RuntimeError(f"No WSGI app found in {module_path}")


def parse_request(stdin: io.TextIOBase) -> Tuple[Dict[str, str], bytes]:
    lines = []
    while True:
        line = stdin.readline()
        if not line or line.strip() == '':
            break
        lines.append(line.rstrip('\n'))

    if not lines:
        return {}, b''

    method, path, *rest = lines[0].split()
    headers = {}
    for line in lines[1:]:
        if ':' in line:
            k, v = line.split(':', 1)
            headers[k.strip()] = v.strip()

    body = sys.stdin.buffer.read()

    environ = {
        'REQUEST_METHOD': method,
        'PATH_INFO': path,
        'SERVER_NAME': 'localhost',
        'SERVER_PORT': '0',
        'wsgi.version': (1, 0),
        'wsgi.url_scheme': 'http',
        'wsgi.input': io.BytesIO(body),
        'wsgi.errors': sys.stderr,
        'wsgi.multithread': False,
        'wsgi.multiprocess': False,
        'wsgi.run_once': True,
        'CONTENT_LENGTH': str(len(body)),
    }
    for k, v in headers.items():
        key = 'HTTP_' + k.upper().replace('-', '_')
        environ[key] = v

    return environ, body


def run_wsgi_app(app, environ: dict):
    response_status = None
    response_headers = []

    def start_response(status, headers, exc_info=None):
        nonlocal response_status, response_headers
        response_status = status
        response_headers = headers
        return sys.stdout.buffer.write

    result = app(environ, start_response)
    body = b''.join(result)
    
    sys.stdout.write("HTTP/1.1 ")
    sys.stdout.write(response_status + '\n')
    for k, v in response_headers:
        sys.stdout.write(f"{k}: {v}\n")
    sys.stdout.write('\n')
    sys.stdout.flush()
    sys.stdout.buffer.write(body)
    sys.stdout.buffer.flush()


def main():
    module_path = sys.argv[1]
    app = load_wsgi_app(module_path)

    environ, _ = parse_request(sys.stdin)
    run_wsgi_app(app, environ)


if __name__ == '__main__':
    main()
