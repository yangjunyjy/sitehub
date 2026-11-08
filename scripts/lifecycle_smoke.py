"""Verify Linux deregistration-before-drain using an isolated real Nacos server."""
from concurrent.futures import ThreadPoolExecutor
import json
import os
from pathlib import Path
import subprocess
import sys
import threading
from urllib.error import HTTPError
from urllib.request import urlopen
import uuid

from governance_smoke import Harness, free_port

root = Path(__file__).resolve().parents[1]
run = root / ".tmp/smoke/runs" / ("lifecycle-" + uuid.uuid4().hex)
run.mkdir(parents=True)
sys.stdout.reconfigure(encoding="utf-8")
harness = Harness(root, run, {})
process = None
log = None
try:
    binary = run / "lifecycle-api"
    subprocess.run(["go", "build", "-o", str(binary), "./tests/e2e/lifecycle"], cwd=root,
                   env={**os.environ, "GOOS": "linux", "GOARCH": "amd64", "CGO_ENABLED": "0"}, check=True)
    harness.start()
    seed_env = {**os.environ, "NACOS_URL": f"http://127.0.0.1:{harness.nacos_port}",
                "NACOS_NAMESPACE_ID": harness.namespace, "NACOS_USERNAME": "nacos",
                "NACOS_PASSWORD": harness.password, "NACOS_SEED_DIR": str(root / "deploy/nacos/config")}
    for iteration in range(2):
        initialized = subprocess.run([sys.executable, str(root / "deploy/nacos/init.py")], env=seed_env, capture_output=True, encoding="utf-8")
        assert initialized.returncode == 0, initialized.stderr
        if iteration == 0:
            harness.publish_config("blog-api", {"Runtime": {"Maintenance": True}})
    preserved = json.loads(harness.nacos("GET", "/cs/configs", {"tenant": harness.namespace, "group": "SITEHUB", "dataId": "blog-api.yaml"}))
    assert preserved["Runtime"]["Maintenance"] is True, "initializer overwrote edited configuration"
    print("PASS Nacos initializer preserves existing configuration", flush=True)
    port = free_port()
    wsl = ["wsl", "-d", os.environ.get("SMOKE_WSL_DISTRO", "Ubuntu"), "-u", "root", "--exec"] if os.name == "nt" else []
    if os.name == "nt":
        ip = subprocess.check_output(wsl + ["hostname", "-I"], encoding="utf-8").split()[0]
    else:
        ip = os.environ["SMOKE_ADVERTISE_IP"]
    config = {
        "Name": "lifecycle-api", "Host": "0.0.0.0", "Port": port, "Mode": "test",
        "Nacos": {"Servers": [f"127.0.0.1:{harness.nacos_port}"], "NamespaceID": harness.namespace,
                  "Group": "SITEHUB", "Username": "nacos", "Password": harness.password,
                  "StateDir": harness.mount_path(run / "client")},
        "Registry": {"Enabled": True, "AdvertiseIP": ip},
        "Shutdown": {"WrapUpTime": "10s", "WaitTime": "20s"},
    }
    path = run / "service.json"
    path.write_text(json.dumps(config), encoding="utf-8")
    log = (run / "service.log").open("wb")
    process = subprocess.Popen(wsl + [harness.mount_path(binary), "-f", harness.mount_path(path)], stdout=log, stderr=log,
                               creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
    def providers():
        return json.loads(harness.nacos("GET", "/ns/instance/list", {"namespaceId": harness.namespace,
                          "groupName": "SITEHUB", "serviceName": "lifecycle-api", "healthyOnly": "true"}))["hosts"]
    harness.wait(lambda: len(providers()) == 1, "lifecycle fixture registration")
    origin = f"http://{ip}:{port}"
    def ready():
        with urlopen(origin + "/readyz", timeout=2) as response:
            return response.status == 200
    harness.wait(ready, "lifecycle HTTP readiness")
    with urlopen(origin + "/test/pid", timeout=3) as response:
        pid = int(response.read())
    started = threading.Event()
    def stream():
        with urlopen(origin + "/test/stream", timeout=25) as response:
            first = response.readline()
            started.set()
            return first + response.read()
    with ThreadPoolExecutor(max_workers=1) as pool:
        inflight = pool.submit(stream)
        assert started.wait(5), "stream did not begin"
        subprocess.run(wsl + ["kill", "-TERM", str(pid)], check=True)
        def unready():
            try:
                with urlopen(origin + "/readyz", timeout=2) as response:
                    return response.status == 503
            except HTTPError as error:
                return error.code == 503
        harness.wait(unready, "readiness withdrawal", 5)
        harness.wait(lambda: len(providers()) == 0, "deregistration before listener closes", 5)
        assert process.poll() is None, "process stopped before inflight drain"
        assert inflight.result(timeout=20).endswith(b"complete\n"), "inflight stream truncated"
    assert process.wait(timeout=15) == 0, "process did not exit gracefully"
    (run / "result.json").write_text(json.dumps({"passed": True, "checks": ["ready before registration", "readiness withdrawn on SIGTERM", "deregistered before drain", "inflight stream completed", "clean process exit"]}, indent=2), encoding="utf-8")
    print("PASS Linux registration, SIGTERM withdrawal, deregistration and stream drain", flush=True)
finally:
    if process and process.poll() is None:
        process.terminate()
        process.wait(timeout=10)
    if log:
        log.close()
    harness.close()
    print("Artifacts:", run, flush=True)
