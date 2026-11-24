"""Start the built stack on an isolated Actions runner and probe its public interfaces."""
import json
import subprocess
import time
from contextlib import ExitStack
from pathlib import Path
from urllib.error import URLError
from urllib.request import Request, urlopen

ROOT = Path(__file__).resolve().parents[1]


def fetch(path, **kwargs):
    with urlopen(Request("http://127.0.0.1:" + path, **kwargs), timeout=4) as response:
        return response.read()


def ready(path, processes):
    for _ in range(90):
        if any(p.poll() is not None for p in processes):
            raise RuntimeError(f"A service exited while checking {path}")
        try:
            return fetch(path)
        except (URLError, TimeoutError):
            time.sleep(1)
    raise RuntimeError(f"Startup timed out at {path}")


def stop(process):
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()


def main():
    with ExitStack() as cleanup:
        processes = []

        def start(cwd, *cmd):
            p = subprocess.Popen(cmd, cwd=ROOT / cwd)
            cleanup.callback(stop, p)
            processes.append(p)

        for service in ("erp", "wms", "scm", "mes"):
            start(service, ROOT / ".bin" / service)
        start("mes", ROOT / ".bin/mes-listener")
        start("mes", ROOT / ".bin/mes-socket")
        start("newGoService/demo", ROOT / ".bin/go-demo")
        start("yq-tablehandle-system/backend", ROOT / ".bin/table-agent")
        start("yq-tablehandle-system/frontend", "node_modules/.bin/vite", "preview",
              "--host", "127.0.0.1", "--port", "5175", "--strictPort")
        war, = (ROOT / "hello-api/target").glob("*.war")
        start("hello-api", "java", "-jar", war)
        for port, service in enumerate(("erp", "wms", "scm", "mes"), 8001):
            assert json.loads(ready(f"{port}/health", processes))["data"]["service"] == service
        assert json.loads(ready("8005/health", processes))["status"] == "ok"
        assert b"go_demo" in ready("8000/hello", processes)
        assert b'id="root"' in ready("8004/", processes)
        assert b'id="root"' in ready("5175/", processes)
        assert json.loads(ready("5175/api/health", processes))["status"] == "ok"
        assert fetch("5175/api/sample").startswith(b"PK")
        assert json.loads(ready("8087/api/products?code=P001", processes))["productCode"] == "P001"
        # Exercise the real stdio transport and tool implementation against the live MES.
        subprocess.run(["node", "--input-type=module", "-e", """
          import { Client } from '@modelcontextprotocol/sdk/client/index.js';
          import { StdioClientTransport } from '@modelcontextprotocol/sdk/client/stdio.js';
          const client = new Client({name:'actions', version:'1'});
          await client.connect(new StdioClientTransport({command:'node',args:['dist/index.js']}));
          try {
            const res = await client.callTool({name:'list_orders',arguments:{}});
            if (res.isError || !res.content.length) throw Error(JSON.stringify(res));
          } finally { await client.close(); }
        """], cwd=ROOT / "mes-orders-mcp", check=True, timeout=25)
        print("Verified HTTP services, both frontends, MES listener, socket and MCP", flush=True)


if __name__ == "__main__":
    main()
