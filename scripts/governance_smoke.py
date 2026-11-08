"""Real native Go gateway/Nacos/etcd integration harness used by smoke.py --governance.

Containers, network and credentials are unique per run; existing deployments
are never stopped or reconfigured. On Windows, Docker runs in WSL Ubuntu.
"""
import base64
from concurrent.futures import ThreadPoolExecutor
import copy
import json
import os
from pathlib import Path
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import secrets
import socket
import subprocess
import threading
import time
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import Request, urlopen


def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


class Harness:
    def __init__(self, root, run, ports):
        self.root, self.run, self.ports = root, run, ports
        self.name = "sitehub-it-" + secrets.token_hex(5)
        self.namespace = self.name
        self.key, self.password = secrets.token_hex(24), secrets.token_hex(24)
        self.token = ""
        self.containers, self.web, self.logs = [], [], []
        self.replicas = []
        self.gateways = []
        self.gateway_addresses = []
        self.http_fixtures = []
        self.advertise_ip = "127.0.0.1"
        self.etcd_port = free_port()
        self.network_created = False
        self.wsl_keepalive = None
        self.gateway_port, self.admin_port = free_port(), free_port()
        while True:
            candidate = free_port()
            if candidate > 64000:
                continue
            try:
                with socket.socket() as sock:
                    sock.bind(("127.0.0.1", candidate + 1000))
                self.nacos_port = candidate
                break
            except OSError:
                pass
        self.nacos_url = f"http://127.0.0.1:{self.nacos_port}/nacos/v1"
        self.admin_url = f"http://127.0.0.1:{self.admin_port}"
        self.gateway_url = f"http://127.0.0.1:{self.gateway_port}"
        self.docker_prefix = ["wsl", "-d", os.environ.get("SMOKE_WSL_DISTRO", "Ubuntu"), "-u", "root", "--exec", "docker"] if os.name == "nt" else ["docker"]
        self.host_ip = os.environ.get("SMOKE_ADVERTISE_IP", "")
        if not self.host_ip:
            if os.name == "nt":
                output = subprocess.check_output(["wsl", "-d", os.environ.get("SMOKE_WSL_DISTRO", "Ubuntu"), "--exec", "ip", "-json", "route"], encoding="utf-8")
                self.host_ip = next(route["gateway"] for route in json.loads(output) if route.get("dst") == "default")
            else:
                self.host_ip = "host.docker.internal"
        self.binary = root / ".tmp/smoke/bin" / ("gatewayctl.exe" if os.name == "nt" else "gatewayctl")

    def docker(self, *args, check=True):
        result = subprocess.run(self.docker_prefix + list(args), capture_output=True, encoding="utf-8", errors="replace", timeout=120)
        if check and result.returncode:
            message = result.stderr.replace(self.key, "[redacted]").replace(self.password, "[redacted]")
            raise RuntimeError("Docker integration command failed: " + message[-1500:])
        return result.stdout.strip()

    def mount_path(self, path):
        if os.name != "nt":
            return str(path)
        return subprocess.check_output(["wsl", "-d", os.environ.get("SMOKE_WSL_DISTRO", "Ubuntu"), "--exec", "wslpath", "-u", str(path)], encoding="utf-8").strip()

    def nacos(self, method, path, values=None, expected=200):
        values = dict(values or {})
        if self.token:
            values["accessToken"] = self.token
        encoded = urlencode(values).encode()
        url = self.nacos_url + path
        if method == "GET":
            url += "?" + encoded.decode()
        request = Request(url, data=None if method == "GET" else encoded, method=method)
        try:
            with urlopen(request, timeout=8) as response:
                status, data = response.status, response.read().decode()
        except HTTPError as error:
            status, data = error.code, error.read().decode()
        if status != expected:
            raise RuntimeError(f"Nacos {method} {path}: {status}, expected {expected}")
        return data

    def wait(self, test, description, timeout=90):
        deadline = time.monotonic() + timeout
        last = None
        while time.monotonic() < deadline:
            try:
                if test():
                    return
            except Exception as error:
                last = type(error).__name__
            time.sleep(.3)
        raise RuntimeError(f"Timed out: {description}; last error type: {last}")

    def start_container(self, suffix, image, options, command):
        name = self.name + "-" + suffix
        self.docker("run", "-d", "--name", name, "--network", self.name, "--network-alias", suffix, *options, image, *command)
        self.containers.append(name)

    def start(self):
        if os.name == "nt":
            # WSL can shut down when its last foreground Windows session exits,
            # even with detached containers. Keep stdin open for this test only.
            self.wsl_keepalive = subprocess.Popen(["wsl", "-d", os.environ.get("SMOKE_WSL_DISTRO", "Ubuntu"), "--exec", "cat"], stdin=subprocess.PIPE, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
        self.docker("network", "create", self.name)
        self.network_created = True
        print("Starting isolated Nacos, etcd and native Go gateway", flush=True)
        self.start_container("nacos", os.environ.get("SMOKE_NACOS_IMAGE", "nacos/nacos-server:v2.5.1"), [
            "-p", f"127.0.0.1:{self.nacos_port}:8848", "-p", f"127.0.0.1:{self.nacos_port + 1000}:9848",
            "-e", "MODE=standalone", "-e", "NACOS_AUTH_ENABLE=true",
            "-e", "NACOS_AUTH_TOKEN=" + base64.b64encode(secrets.token_bytes(48)).decode(),
            "-e", "NACOS_AUTH_IDENTITY_KEY=sitehub-it", "-e", "NACOS_AUTH_IDENTITY_VALUE=" + secrets.token_hex(24),
            "-e", "JVM_XMS=256m", "-e", "JVM_XMX=512m", "-e", "JVM_XMN=128m",
        ], [])
        self.start_container("gateway-etcd", "quay.io/coreos/etcd:v3.5.21", ["-p", f"127.0.0.1:{self.etcd_port}:2379"], [
            "etcd", "--listen-client-urls=http://0.0.0.0:2379", "--advertise-client-urls=http://gateway-etcd:2379",
        ])
        self.wait(lambda: self.nacos("GET", "/console/health/readiness"), "Nacos readiness", 120)
        self.nacos("POST", "/auth/users/admin", {"password": self.password})
        self.token = json.loads(self.nacos("POST", "/auth/login", {"username": "nacos", "password": self.password}))["accessToken"]
        self.nacos("POST", "/console/namespaces", {"customNamespaceId": self.namespace, "namespaceName": self.namespace, "namespaceDesc": "isolated integration test"})
        self.seed_manifest = self.run / "gateway-seed.json"
        self.seed_manifest.write_text((self.root / "gateway/etc/routes.json").read_text(encoding="utf-8"), encoding="utf-8")
        self.start_gateway("gateway", self.gateway_port, self.admin_port)
        self.ctl_config = self.run / "gatewayctl.json"
        self.ctl_config.write_text(json.dumps({"AdminURL": self.admin_url, "AdminKey": self.key, "StateDir": str(self.run / "releases")}), encoding="utf-8")

    def start_gateway(self, name, public_port, admin_port):
        config = {
            "Name": name, "Listen": f"127.0.0.1:{public_port}", "Manifest": str(self.seed_manifest),
            "Log": {"Mode": "console", "Level": "error"},
            "Admin": {"Listen": f"127.0.0.1:{admin_port}", "Key": self.key, "AllowedCIDRs": ["127.0.0.1/32"]},
            "Etcd": {"Endpoints": [f"http://127.0.0.1:{self.etcd_port}"], "Key": f"/sitehub/{self.name}/gateway/v1/config"},
            "Nacos": {"Servers": [f"127.0.0.1:{self.nacos_port}"], "NamespaceID": self.namespace, "Group": "SITEHUB", "Username": "nacos", "Password": self.password, "StateDir": str(self.run / (name + "-nacos"))},
            "Discovery": {"Enabled": True, "MaxStale": "5s"},
            "HTTP": {"ShutdownTimeout": "5s"},
        }
        path = self.run / (name + "-bootstrap.json")
        path.write_text(json.dumps(config), encoding="utf-8")
        log = (self.run / (name + ".log")).open("wb")
        self.logs.append(log)
        binary = self.root / ".tmp/smoke/bin" / ("gateway.exe" if os.name == "nt" else "gateway")
        process = subprocess.Popen([str(binary), "-f", str(path)], cwd=self.root, stdout=log, stderr=log, creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
        self.gateways.append(process)
        def ready():
            if process.poll() is not None:
                raise RuntimeError("native gateway exited; see " + str(self.run / (name + ".log")))
            with urlopen(f"http://127.0.0.1:{admin_port}/readyz", timeout=3) as response:
                return response.status == 200
        self.wait(ready, "native gateway readiness " + name)
        self.gateway_addresses.append(f"http://127.0.0.1:{public_port}")
        return f"http://127.0.0.1:{public_port}", f"http://127.0.0.1:{admin_port}"

    def control(self, method="GET", path="/v1/config", data=None, revision=None, admin_url=None, expected=200):
        headers = {"Authorization": "Bearer " + self.key, "Content-Type": "application/json"}
        if revision is not None:
            headers["If-Match"] = '"' + str(revision) + '"'
        request = Request((admin_url or self.admin_url) + path, headers=headers, method=method,
                          data=json.dumps(data).encode() if data is not None else None)
        try:
            with urlopen(request, timeout=15) as response:
                status, body = response.status, response.read()
        except HTTPError as error:
            status, body = error.code, error.read()
        if expected is not None and status != expected:
            raise AssertionError(f"gateway control {method} {path}: {status}, expected {expected}")
        value = json.loads(body)
        return (status, value) if expected is None else value

    def publish_config(self, name, data):
        content = data if isinstance(data, str) else json.dumps(data)
        self.nacos("POST", "/cs/configs", {"tenant": self.namespace, "group": "SITEHUB", "dataId": name + ".yaml", "content": content, "type": "yaml"})

    def configure(self, name, data):
        if name not in [f"{d}-{k}" for d in ["rbac", "blog", "product", "file"] for k in ["api", "rpc"]] + ["scheduler"]:
            return data
        data = copy.deepcopy(data)
        for value in data.values():
            if isinstance(value, dict) and "Endpoints" in value:
                endpoint = value.pop("Endpoints")[0]
                target = next(key for key, port in self.ports.items() if str(port) == endpoint.rsplit(":", 1)[1])
                value["Target"] = "nacos:///" + target
        data.update(Nacos={"Servers": [f"127.0.0.1:{self.nacos_port}"], "NamespaceID": self.namespace, "Group": "SITEHUB", "Username": "nacos", "Password": self.password, "StateDir": str(self.run / "nacos-client")},
                    Registry={"Enabled": True, "AdvertiseIP": self.advertise_ip, "AllowLoopback": True, "MaxStale": "5s", "StartupTimeout": "30s"},
                    ConfigCenter={"Enabled": True, "DataID": name + ".yaml"},
                    Runtime={"Maintenance": False, "MaxConcurrentRequests": 1000},
                    Shutdown={"WrapUpTime": "10s", "WaitTime": "75s"})
        if "ListenOn" in data:
            data["ListenOn"] = "0.0.0.0:" + data["ListenOn"].rsplit(":", 1)[1]
        if "Host" in data:
            data["Host"] = "0.0.0.0"
            data["TrustedProxyCIDRs"] = ["127.0.0.1/32"]
        if name == "rbac-api":
            data["BrowserSession"] = {"Secure": False}
        self.publish_config(name, {"Runtime": data["Runtime"]})
        return data

    def ctl(self, action, manifest=None, release=None):
        args = [str(self.binary), action, "-f", str(self.ctl_config)]
        if manifest:
            args.extend(["--manifest", str(manifest)])
        if release:
            args.extend(["--release", str(release)])
        result = subprocess.run(args, cwd=self.root, capture_output=True, encoding="utf-8", errors="replace", timeout=60)
        with (self.run / "gatewayctl.log").open("a", encoding="utf-8") as output:
            output.write(result.stdout + result.stderr)
        if result.returncode:
            raise RuntimeError("gatewayctl failed: " + result.stderr[-1000:])

    def request(self, path, host="sitehub.local", headers=None, method="GET", body=None, gateway_url=None):
        request = Request((gateway_url or self.gateway_url) + path, headers={"Host": host, **(headers or {})}, method=method, data=body)
        try:
            with urlopen(request, timeout=10) as response:
                return response.status, response.read(), dict(response.headers)
        except HTTPError as error:
            return error.code, error.read(), dict(error.headers)

    def setup_routes(self):
        if hasattr(self, "manifest"):
            return
        manifest = json.loads((self.root / "gateway/etc/routes.json").read_text(encoding="utf-8"))
        node_ports = {}
        for app in ["site", "admin"]:
            port = free_port()
            node_ports[app] = port
            log = (self.run / ("web-" + app + ".log")).open("wb")
            self.logs.append(log)
            self.web.append(subprocess.Popen(["node", str(self.root / "web/server/start.mjs"), "--app", app], cwd=self.root,
                           env={**os.environ, "HOST": "0.0.0.0", "PORT": str(port)}, stdout=log, stderr=log, creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0)))
            def ready():
                with urlopen(f"http://127.0.0.1:{port}/healthz", timeout=2) as response:
                    return response.status == 200
            self.wait(ready, "Node " + app)
        for name, upstream in manifest["upstreams"].items():
            if name.startswith("web-"):
                upstream["nodes"] = [{"url": f"http://127.0.0.1:{node_ports[name[4:]]}", "weight": 1}]
        for audience, plugins in manifest["profiles"].items():
            host = "admin.sitehub.local" if audience == "admin" else "sitehub.local"
            plugins["browser-session"]["origins"].append(f"http://{host}:{self.gateway_port}")
        self.manifest = self.run / "gateway.json"
        self.manifest.write_text(json.dumps(manifest), encoding="utf-8")
        self.ctl("validate", self.manifest)
        self.ctl("publish", self.manifest)
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 200, "gateway to Nacos-discovered blog API")

    def verify(self, member_token):
        def instances(name):
            return json.loads(self.nacos("GET", "/ns/instance/list", {"namespaceId": self.namespace, "groupName": "SITEHUB", "serviceName": name, "healthyOnly": "true"}))["hosts"]
        for name in self.ports:
            self.wait(lambda: len(instances(name)) == 1, "registration of " + name)
        assert not instances("scheduler"), "scheduler registered a fake provider"
        self.setup_routes()
        assert self.request("/", "unknown.test")[0] == 404
        assert self.request("/v1/unknown/route")[0] == 404
        assert self.request("/internal/governance")[0] != 200 or b"appliedVersion" not in self.request("/internal/governance")[1]
        assert self.request("/v1/blog/articles", headers={"Cookie": "sitehub_site_session=a; sitehub_site_session=b"})[0] == 400
        assert self.request("/v1/rbac/members/logout", headers={"Cookie": "sitehub_site_session=fake", "Origin": "https://foreign.test"}, method="POST", body=b"{}")[0] == 403
        ranged = self.request("/v1/file/download?id=native", headers={"Cookie": "sitehub_site_session=" + member_token, "Range": "bytes=2-5"})
        assert ranged[0] == 206 and ranged[1] == b"2345"
        assert self.request("/v1/file/download?id=native", host="admin.sitehub.local", headers={"Cookie": "sitehub_site_session=" + member_token})[0] == 401
        assert self.request("/v1/file/download?id=native", headers={"Cookie": "sitehub_site_session=invalid", "Authorization": "Bearer " + member_token}, method="HEAD")[0] == 200
        print("GOVERNANCE PASS registry, gateway discovery, host routing and CSRF", flush=True)
        direct = f"http://127.0.0.1:{self.ports['blog-api']}"
        def status():
            with urlopen(direct + "/internal/governance", timeout=3) as response:
                return json.load(response)
        initial = status()["appliedVersion"]
        self.publish_config("blog-api", {"Runtime": {"Maintenance": True, "MaxConcurrentRequests": 10}})
        self.wait(lambda: status()["appliedVersion"] != initial, "dynamic config commit")
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 503, "maintenance through gateway")
        applied = status()["appliedVersion"]
        self.publish_config("blog-api", {"Runtime": {"MaxConcurrentRequests": -1}})
        self.wait(lambda: bool(status().get("lastError")), "invalid config rejected")
        assert status()["appliedVersion"] == applied
        self.publish_config("blog-api", {"Port": 9999, "Runtime": {"Maintenance": False}})
        self.wait(lambda: "restart required" in status().get("lastError", ""), "static config rejected")
        assert status()["appliedVersion"] == applied
        self.publish_config("blog-api", {"Runtime": {"Maintenance": False, "MaxConcurrentRequests": 1000}})
        self.wait(lambda: not status().get("lastError"), "config recovery")
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 200, "traffic recovery", 50)
        print("GOVERNANCE PASS config commit, rejection, restart boundary and recovery", flush=True)
        updated = json.loads(self.manifest.read_text())
        updated["routes"].append({"id": "site-rate-probe", "host": "sitehub.local", "paths": ["/v1/blog/articles"], "priority": 300,
                                  "upstream": "blog", "profile": "site", "plugins": {"limit-count": {"count": 1, "window_seconds": 60}}})
        for route in updated["routes"]:
            if route["id"] == "site-unknownall":
                route["plugins"]["static-response"]["status"] = 410
        changed = self.run / "gateway-changed.json"
        changed.write_text(json.dumps(updated), encoding="utf-8")
        before = set((self.run / "releases").glob("*.json"))
        self.ctl("publish", changed)
        release = next(iter(set((self.run / "releases").glob("*.json")) - before))
        self.wait(lambda: self.request("/v1/unknown/route")[0] == 410, "route hot update")
        assert self.request("/v1/blog/articles")[0] == 200
        assert self.request("/v1/blog/articles")[0] == 429
        self.ctl("rollback", release=release)
        self.wait(lambda: self.request("/v1/unknown/route")[0] == 404, "route rollback")
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 200, "removed limit route")
        print("GOVERNANCE PASS gateway hot publish and versioned rollback", flush=True)
        self.verify_shared_configuration()
        self.verify_http_discovery()
        self.verify_discovery_recovery()
        (self.run / "governance-result.json").write_text(json.dumps({"passed": True, "checks": [
            "eight service registrations and scheduler consumer only",
            "Host routing, audience isolation, Origin checks and Range/HEAD",
            "business runtime config commit, invalid/static rejection and recovery",
            "gateway whole-document publish, rate limit and versioned rollback",
            "two gateway replicas, concurrent CAS and configuration convergence",
            "etcd outage retains applied traffic and watch resumes after recovery",
            "HTTP authoritative empty discovery list and recovery",
            "RPC replica failover, authoritative empty discovery and recovery",
            "Nacos outage expires stale HTTP/RPC addresses and automatically recovers",
        ]}, indent=2), encoding="utf-8")

    def verify_shared_configuration(self):
        public_port, admin_port = free_port(), free_port()
        replica_url, replica_admin = self.start_gateway("gateway-replica", public_port, admin_port)
        before = self.control()
        self.wait(lambda: self.control(admin_url=replica_admin)["applied_revision"] == before["revision"], "second gateway applies current configuration")
        self.wait(lambda: self.request("/v1/blog/articles", gateway_url=replica_url)[0] == 200, "second gateway discovers API services")

        def change_status(document, status):
            document = copy.deepcopy(document)
            for route in document["routes"]:
                if route["id"] in ("site-unknownall", "admin-unknownall"):
                    route["plugins"]["static-response"]["status"] = status
            return document

        # Both writers deliberately start from the same observed revision. The
        # store must accept one complete candidate and reject the stale writer.
        candidates = [change_status(before["manifest"], status) for status in (410, 418)]
        with ThreadPoolExecutor(max_workers=2) as pool:
            futures = [pool.submit(self.control, "PUT", data=document, revision=before["revision"], admin_url=admin, expected=None)
                       for document, admin in zip(candidates, (self.admin_url, replica_admin))]
            outcomes = [future.result() for future in futures]
        assert sorted(status for status, _ in outcomes) == [200, 409], "concurrent publishers did not enforce CAS"
        winner = next(value for status, value in outcomes if status == 200)
        winning_status = next(route["plugins"]["static-response"]["status"] for route in winner["manifest"]["routes"] if route["id"] == "site-unknownall")
        for admin in (self.admin_url, replica_admin):
            self.wait(lambda admin=admin: self.control(admin_url=admin)["applied_revision"] == winner["revision"], "gateway revision convergence")
        for gateway in (self.gateway_url, replica_url):
            assert self.request("/v1/unknown/route", gateway_url=gateway)[0] == winning_status
            assert self.request("/v1/unknown/route", host="admin.sitehub.local", gateway_url=gateway)[0] == winning_status

        invalid = copy.deepcopy(winner["manifest"])
        invalid["routes"][0].setdefault("plugins", {})["unknown-plugin"] = {}
        self.control("PUT", data=invalid, revision=winner["revision"], admin_url=replica_admin, expected=400)
        assert self.control()["revision"] == winner["revision"], "invalid candidate changed the stored revision"
        restored = self.control("PUT", data=before["manifest"], revision=winner["revision"])
        self.wait(lambda: self.control(admin_url=replica_admin)["applied_revision"] == restored["revision"], "restored config reaches replica")
        print("GOVERNANCE PASS two gateways, whole-document CAS, invalid candidate rejection and revision convergence", flush=True)

        # Existing traffic does not depend on the configuration store's live
        # availability. A failed publication must not alter the applied snapshot.
        self.docker("stop", "--time", "5", self.name + "-gateway-etcd")
        self.control(expected=503)
        for gateway in (self.gateway_url, replica_url):
            assert self.request("/v1/blog/articles", gateway_url=gateway)[0] == 200
            assert self.request("/v1/unknown/route", gateway_url=gateway)[0] == 404
        self.control("PUT", data=candidates[0], revision=restored["revision"], expected=503)
        assert self.request("/v1/unknown/route")[0] == 404
        self.docker("start", self.name + "-gateway-etcd")
        self.wait(lambda: self.control()["revision"] == restored["revision"], "etcd preserves last committed config after restart")
        next_state = self.control("PUT", data=change_status(before["manifest"], 410), revision=restored["revision"], admin_url=replica_admin)
        self.wait(lambda: self.control()["applied_revision"] == next_state["revision"], "watch resumes after etcd restart")
        assert self.request("/v1/unknown/route")[0] == 410
        final_state = self.control("PUT", data=before["manifest"], revision=next_state["revision"])
        self.wait(lambda: self.control(admin_url=replica_admin)["applied_revision"] == final_state["revision"], "final rollback after etcd recovery")
        assert self.request("/v1/unknown/route", gateway_url=replica_url)[0] == 404
        print("GOVERNANCE PASS etcd outage retains applied traffic, rejects uncommitted changes and resumes watch", flush=True)

    def verify_http_discovery(self):
        def enable(enabled):
            self.nacos("PUT", "/ns/instance", {"namespaceId": self.namespace, "groupName": "SITEHUB", "serviceName": "blog-api",
                       "ip": self.advertise_ip, "port": self.ports["blog-api"], "clusterName": "DEFAULT", "ephemeral": "true", "enabled": str(enabled).lower(), "weight": 100})
        enable(False)
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 503, "authoritative empty HTTP API list")
        # The API still serves directly: 503 must originate from registry
        # withdrawal, not an unavailable API or its downstream RPC service.
        with urlopen(f"http://127.0.0.1:{self.ports['blog-api']}/health", timeout=3) as response:
            assert response.status == 200
        enable(True)
        self.wait(lambda: self.request("/v1/blog/articles")[0] == 200, "HTTP discovery and health recovery")
        print("GOVERNANCE PASS HTTP disabled instance, authoritative empty list and recovery", flush=True)

    def verify_discovery_recovery(self):
        replica_port = free_port()
        config = json.loads((self.run / "blog-rpc.json").read_text())
        config["ListenOn"] = f"0.0.0.0:{replica_port}"
        config["Nacos"]["StateDir"] = str(self.run / "replica-cache")
        path = self.run / "blog-rpc-replica.json"
        path.write_text(json.dumps(config), encoding="utf-8")
        log = (self.run / "blog-rpc-replica.log").open("wb")
        self.logs.append(log)
        binary = self.root / ".tmp/smoke/bin" / ("blog-rpc.exe" if os.name == "nt" else "blog-rpc")
        process = subprocess.Popen([str(binary), "-f", str(path)], cwd=self.root, stdout=log, stderr=log, creationflags=getattr(subprocess, "CREATE_NO_WINDOW", 0))
        self.replicas.append(process)
        def providers():
            return json.loads(self.nacos("GET", "/ns/instance/list", {"namespaceId": self.namespace, "groupName": "SITEHUB", "serviceName": "blog-rpc", "healthyOnly": "true"}))["hosts"]
        self.wait(lambda: len(providers()) == 2, "second RPC instance")
        def enable(port, enabled):
            self.nacos("PUT", "/ns/instance", {"namespaceId": self.namespace, "groupName": "SITEHUB", "serviceName": "blog-rpc", "ip": self.advertise_ip, "port": port, "clusterName": "DEFAULT", "ephemeral": "true", "enabled": str(enabled).lower(), "weight": 100})
        def health():
            try:
                with urlopen(f"http://127.0.0.1:{self.ports['blog-api']}/health", timeout=3) as response:
                    return response.status
            except HTTPError as error:
                return error.code
        enable(self.ports["blog-rpc"], False)
        self.wait(lambda: health() == 200, "RPC failover to second instance")
        enable(replica_port, False)
        self.wait(lambda: health() == 503, "authoritative empty RPC list")
        enable(self.ports["blog-rpc"], True)
        self.wait(lambda: health() == 200, "RPC recovery after empty list")
        process.terminate()
        process.wait(timeout=10)
        print("GOVERNANCE PASS two RPC instances, disabled instances, empty list and recovery", flush=True)
        def outage_checks():
            self.wait(lambda: health() == 503, "registry disconnect exceeds RPC MaxStale", 60)
            self.wait(lambda: self.request("/v1/blog/articles")[0] == 503, "gateway business requests during registry outage", 60)
        def recovery_checks():
            self.wait(lambda: health() == 200, "RPC re-registration after Nacos restart", 90)
            self.wait(lambda: self.request("/v1/blog/articles")[0] == 200, "gateway business recovery after Nacos restart", 90)
        self.verify_registry_probe(outage_checks, recovery_checks)
        print("GOVERNANCE PASS Nacos outage, stale-address expiry and automatic recovery", flush=True)

    def verify_registry_probe(self, outage_checks=None, recovery_checks=None):
        """Prove HTTP MaxStale independently of API/RPC health dependencies.

        This persistent Nacos instance remains a healthy ordinary HTTP server
        while the registry is stopped. Thus a gateway 503 cannot be attributed
        to the backend's health endpoint failing because of another service.
        """
        class ProbeHandler(BaseHTTPRequestHandler):
            def do_GET(self):
                body = b'{"probe":"healthy"}'
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(body)))
                self.end_headers()
                self.wfile.write(body)

            def log_message(self, *_):
                pass

        server = ThreadingHTTPServer(("0.0.0.0", 0), ProbeHandler)
        server.daemon_threads = True
        worker = threading.Thread(target=server.serve_forever, name="registry-probe", daemon=True)
        worker.start()
        self.http_fixtures.append((server, worker))
        port = server.server_port
        # Persistent Nacos instances are actively checked by the Nacos server,
        # so this address must also be reachable from its Docker container.
        probe_ip = socket.gethostbyname(self.host_ip)
        instance = {"namespaceId": self.namespace, "groupName": "SITEHUB", "serviceName": "probe-api",
                    "ip": probe_ip, "port": port, "clusterName": "DEFAULT", "ephemeral": "false", "enabled": "true", "healthy": "true", "weight": 100}
        self.nacos("POST", "/ns/instance", instance)
        before = self.control()
        document = copy.deepcopy(before["manifest"])
        document["upstreams"]["registry-probe"] = {
            "service": "probe-api", "connect_timeout_ms": 1000, "response_header_timeout_ms": 3000, "io_timeout_ms": 3000,
            "health_path": "/readyz", "health_interval_ms": 200, "health_timeout_ms": 500, "health_failures": 1,
        }
        document["routes"].append({"id": "registry-probe", "host": "sitehub.local", "paths": ["/registry-probe"], "priority": 500,
                                   "upstream": "registry-probe", "plugins": {"request-id": {}}})
        self.control("PUT", data=document, revision=before["revision"])

        def direct_health():
            with urlopen(f"http://127.0.0.1:{port}/readyz", timeout=2) as response:
                return response.status == 200 and json.load(response) == {"probe": "healthy"}

        for gateway in self.gateway_addresses:
            self.wait(lambda gateway=gateway: self.request("/registry-probe", gateway_url=gateway)[0] == 200, "gateway discovers standalone healthy HTTP probe")
        assert direct_health()
        self.nacos("PUT", "/ns/instance", {**instance, "enabled": "false"})
        for gateway in self.gateway_addresses:
            self.wait(lambda gateway=gateway: self.request("/registry-probe", gateway_url=gateway)[0] == 503, "HTTP probe authoritative empty list")
        assert direct_health(), "probe health changed while disabled in discovery"
        self.nacos("PUT", "/ns/instance", instance)
        for gateway in self.gateway_addresses:
            self.wait(lambda gateway=gateway: self.request("/registry-probe", gateway_url=gateway)[0] == 200, "HTTP probe discovery re-enabled")
        print("GOVERNANCE PASS independently healthy HTTP probe discovery, disabled instance and recovery", flush=True)

        self.docker("stop", "--time", "5", self.name + "-nacos")
        assert direct_health(), "healthy HTTP fixture must survive registry outage"
        for gateway in self.gateway_addresses:
            self.wait(lambda gateway=gateway: self.request("/registry-probe", gateway_url=gateway)[0] == 503, "HTTP discovery MaxStale expires despite healthy upstream", 60)
        assert direct_health(), "503 must come from discovery expiry, not upstream health"
        print("GOVERNANCE PASS HTTP MaxStale returns 503 while direct upstream health remains 200", flush=True)
        if outage_checks:
            outage_checks()
        self.docker("start", self.name + "-nacos")
        self.wait(lambda: self.nacos("GET", "/console/health/readiness"), "restarted Nacos", 120)
        for gateway in self.gateway_addresses:
            self.wait(lambda gateway=gateway: self.request("/registry-probe", gateway_url=gateway)[0] == 200, "HTTP discovery automatically recovers existing persistent instance", 90)
        assert direct_health()
        if recovery_checks:
            recovery_checks()
        current = self.control()
        self.control("PUT", data=before["manifest"], revision=current["revision"])
        self.nacos("DELETE", "/ns/instance", instance)
        server.shutdown()
        server.server_close()
        worker.join(timeout=5)
        self.http_fixtures.remove((server, worker))
        print("GOVERNANCE PASS independent healthy HTTP probe: empty discovery, MaxStale expiry and automatic recovery", flush=True)

    def browser_env(self):
        return {"WEB_ADMIN_URL": f"http://admin.sitehub.local:{self.gateway_port}", "WEB_SITE_URL": f"http://sitehub.local:{self.gateway_port}"}

    def close(self):
        for process in self.gateways + self.web + self.replicas:
            if process.poll() is None:
                process.terminate()
            try:
                process.wait(timeout=10)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
        for server, worker in self.http_fixtures:
            server.shutdown()
            server.server_close()
            worker.join(timeout=5)
        self.http_fixtures.clear()
        for log in self.logs:
            log.close()
        for name in reversed(self.containers):
            logs = self.docker("logs", name, check=False)
            (self.run / (name + ".log")).write_text(logs.replace(self.key, "[redacted]").replace(self.password, "[redacted]"), encoding="utf-8")
            self.docker("rm", "-f", name, check=False)
        if self.network_created:
            self.docker("network", "rm", self.name, check=False)
        if self.wsl_keepalive:
            self.wsl_keepalive.stdin.close()
            try:
                self.wsl_keepalive.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.wsl_keepalive.terminate()
