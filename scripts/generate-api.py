"""Refresh generated routes and DTOs without duplicating grouped HTTP handlers."""
import re
import shutil
import subprocess
import sys
from pathlib import Path
from tempfile import TemporaryDirectory

root = Path(__file__).resolve().parents[1]
domain = sys.argv[1]
relative = Path("service", domain, "api")
api = root / relative

with TemporaryDirectory(prefix="sitehub-goctl-") as temporary:
    stage = Path(temporary)
    shutil.copyfile(root / "go.mod", stage / "go.mod")
    output = stage / relative
    subprocess.run(["goctl", "api", "go", "-api", str(api / f"{domain}.api"),
                    "-dir", str(output), "--style", "go_zero"], cwd=stage, check=True)
    for name in ("types/types.go", "handler/routes.go"):
        shutil.copyfile(output / "internal" / name, api / "internal" / name)

    handlers = api / "internal/handler"
    declared = set()
    for source in handlers.rglob("*.go"):
        package = source.parent.relative_to(handlers)
        for name in re.findall(r"^func (\w+Handler)\(", source.read_text(encoding="utf-8"), re.M):
            declared.add((package, name))
    generated = output / "internal/handler"
    for source in generated.rglob("*_handler.go"):
        package = source.parent.relative_to(generated)
        name = re.search(r"^func (\w+Handler)\(", source.read_text(encoding="utf-8"), re.M)[1]
        if (package, name) in declared:
            continue
        destination = handlers / package / source.name
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(source, destination)
        logic = Path("internal/logic") / package / source.name.replace("_handler.go", "_logic.go")
        if not (api / logic).exists():
            (api / logic).parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(output / logic, api / logic)
