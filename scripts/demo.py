"""Own a dedicated Redis on port 6380 and run the local demo; Ctrl+C cleans up."""
import os
from pathlib import Path
import shutil
import signal
import socket
import subprocess
import sys
import time

ROOT = Path(__file__).resolve().parent.parent


def interrupted(signum, frame):
    raise KeyboardInterrupt


def main():
    if not shutil.which("redis-server"):
        raise SystemExit("redis-server not found. Install Redis or use make dev with REDIS_ADDR.")
    with socket.socket() as sock:
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        try:
            sock.bind(("127.0.0.1", 6380))
        except OSError:
            raise SystemExit("Port 6380 is occupied; stop the previous demo first.")
    data = ROOT / ".data" / "redis"
    data.mkdir(parents=True, exist_ok=True)
    (ROOT / ".logs").mkdir(exist_ok=True)
    processes = []
    signal.signal(signal.SIGINT, interrupted)
    signal.signal(signal.SIGTERM, interrupted)
    with (ROOT / ".logs" / "redis.log").open("w") as logfile:
        try:
            redis = subprocess.Popen([
                "redis-server", "--bind", "127.0.0.1", "--port", "6380",
                "--appendonly", "yes", "--dir", str(data),
            ], stdout=logfile, stderr=subprocess.STDOUT, env=dict(os.environ, LC_ALL="C"))
            processes.append(redis)
            for _ in range(100):
                if redis.poll() is not None:
                    raise RuntimeError("Redis exited; see .logs/redis.log")
                try:
                    with socket.create_connection(("127.0.0.1", 6380), timeout=.2):
                        break
                except OSError:
                    time.sleep(.1)
            else:
                raise RuntimeError("Redis startup timeout")
            env = dict(os.environ, REDIS_ADDR="127.0.0.1:6380", REDIS_PASSWORD="")
            services = subprocess.Popen([sys.executable, str(ROOT / "scripts/services.py"), "run"], env=env)
            processes.append(services)
            while services.poll() is None:
                if redis.poll() is not None:
                    raise RuntimeError("Redis exited; see .logs/redis.log")
                time.sleep(.5)
            return services.returncode
        except KeyboardInterrupt:
            return 0
        finally:
            for process in reversed(processes):
                if process.poll() is None:
                    process.terminate()
                try:
                    process.wait(timeout=20)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()


if __name__ == "__main__":
    sys.exit(main())
