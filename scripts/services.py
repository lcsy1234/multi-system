"""Run local service binaries and check their HTTP endpoints (Python 3 stdlib only)."""

import json
from pathlib import Path
import signal
import socket
import subprocess
import sys
import time
from urllib.error import URLError
from urllib.request import urlopen


ROOT = Path(__file__).resolve().parent.parent
SERVICES = {"erp": 8001, "wms": 8002, "scm": 8003, "mes": 8004}


def get_json(port, path):
    with urlopen(f"http://127.0.0.1:{port}{path}", timeout=1) as response:
        return json.load(response)


def stop_on_signal(signum, frame):
    raise KeyboardInterrupt


def main():
    if len(sys.argv) != 2 or sys.argv[1] not in ("run", "smoke"):
        raise SystemExit("Usage: python3 scripts/services.py run|smoke (or make dev|smoke)")

    # Refuse occupied ports so checks never silently target an existing process.
    for name, port in SERVICES.items():
        with socket.socket() as sock:
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            try:
                sock.bind(("127.0.0.1", port))
            except OSError as error:
                raise SystemExit(f"{name}: port {port} unavailable: {error}")

    processes = {}
    logs = []
    signal.signal(signal.SIGTERM, stop_on_signal)
    signal.signal(signal.SIGINT, stop_on_signal)
    (ROOT / ".logs").mkdir(exist_ok=True)
    try:
        for name in SERVICES:
            logfile = (ROOT / ".logs" / f"{name}.log").open("w")
            logs.append(logfile)
            processes[name] = subprocess.Popen(
                [str(ROOT / ".bin" / name)],
                cwd=ROOT / name,
                stdout=logfile,
                stderr=subprocess.STDOUT,
            )

        deadline = time.monotonic() + 20
        pending = set(SERVICES)
        while pending:
            for name, process in processes.items():
                if process.poll() is not None:
                    raise RuntimeError(f"{name} exited; see .logs/{name}.log")
            for name in list(pending):
                try:
                    result = get_json(SERVICES[name], "/health")
                except (URLError, TimeoutError, OSError):
                    continue
                expected = {"service": name, "status": "ok"}
                if result.get("code") != 0 or any(
                    result.get("data", {}).get(key) != value
                    for key, value in expected.items()
                ):
                    raise RuntimeError(f"{name}: invalid health response: {result}")
                pending.remove(name)
            if time.monotonic() > deadline:
                raise RuntimeError(f"Startup timeout: {sorted(pending)}; see .logs/")
            time.sleep(0.1)

        for name, port in SERVICES.items():
            spec = get_json(port, "/api.json")
            if "get" not in spec.get("paths", {}).get("/health", {}):
                raise RuntimeError(f"{name}: OpenAPI missing GET /health")
            with urlopen(f"http://127.0.0.1:{port}/swagger", timeout=2) as response:
                if response.status != 200:
                    raise RuntimeError(f"{name}: Swagger unavailable")
            print(f"OK {name}: http://127.0.0.1:{port}/health | /api.json | /swagger", flush=True)

        if sys.argv[1] == "run":
            print("MES demo: http://127.0.0.1:8004/ | Logs: .logs/*.log", flush=True)
            print("Listener is NOT started. Run make listener in another terminal to consume queued tasks.", flush=True)
            while True:
                for name, process in processes.items():
                    if process.poll() is not None:
                        raise RuntimeError(f"{name} exited; see .logs/{name}.log")
                time.sleep(0.5)
    except KeyboardInterrupt:
        pass
    finally:
        for process in processes.values():
            if process.poll() is None:
                process.terminate()
        for process in processes.values():
            try:
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
        for logfile in logs:
            logfile.close()


if __name__ == "__main__":
    main()
