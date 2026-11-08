"""Create the development namespace and seed missing configs; never overwrite edits."""
import json
import os
from pathlib import Path
from urllib.error import HTTPError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

BASE = os.environ.get("NACOS_URL", "http://nacos:8848") + "/nacos/v1"
TOKEN = ""


def request(method, path, values=None, missing_ok=False):
    values = dict(values or {})
    if TOKEN:
        values["accessToken"] = TOKEN
    encoded = urlencode(values).encode()
    url = BASE + path
    data = None
    if method == "GET":
        url += "?" + encoded.decode()
    else:
        data = encoded
    try:
        with urlopen(Request(url, data=data, method=method), timeout=10) as response:
            return response.read().decode()
    except HTTPError as error:
        if missing_ok and error.code == 404:
            return None
        raise RuntimeError(f"Nacos {method} {path} failed ({error.code})") from None


def main():
    global TOKEN
    # Nacos 2.4+ requires explicit initialization of the administrator password.
    try:
        request("POST", "/auth/users/admin", {"password": os.environ["NACOS_PASSWORD"]})
    except RuntimeError:
        pass  # Already initialized: the following login must still succeed.
    TOKEN = json.loads(request("POST", "/auth/login", {
        "username": os.environ["NACOS_USERNAME"], "password": os.environ["NACOS_PASSWORD"]
    }))["accessToken"]
    namespace = os.environ["NACOS_NAMESPACE_ID"]
    spaces = json.loads(request("GET", "/console/namespaces"))["data"]
    if not any(space["namespace"] == namespace for space in spaces):
        request("POST", "/console/namespaces", {
            "customNamespaceId": namespace, "namespaceName": namespace,
            "namespaceDesc": "sitehub development environment",
        })
    for path in sorted(Path(os.environ.get("NACOS_SEED_DIR", "/config")).glob("*.yaml")):
        params = {"tenant": namespace, "group": "SITEHUB", "dataId": path.name}
        current = request("GET", "/cs/configs", params, missing_ok=True)
        if current is None:
            request("POST", "/cs/configs", dict(params, content=path.read_text(), type="yaml"))
            print("created", path.name)
        else:
            print("kept", path.name)


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        # Never print a request URL or the SDK/HTTP response containing secrets.
        raise SystemExit(f"Nacos initialization failed: {type(error).__name__}")
