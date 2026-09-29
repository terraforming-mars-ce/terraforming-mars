"""Preserve generated card masters and invoke the shared runtime exporter."""

import json
import subprocess
from io import BytesIO
from pathlib import Path
from PIL import Image
from . import config


def process_and_save(image_data: bytes, output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    image = Image.open(BytesIO(image_data))
    temporary = output_path.with_suffix(".tmp")
    if image.format == "PNG":
        temporary.write_bytes(image_data)
    else:
        image.save(temporary, "PNG")
    temporary.replace(output_path)

    catalog = json.loads(config.CATALOG_PATH.read_text())
    asset_id = f"cards/{output_path.stem}"
    source = f"cards/{output_path.name}"
    previous_source = None
    for entry in catalog["assets"]:
        if entry["id"] == asset_id:
            previous_source = entry["source"]
            entry["source"] = source
            break
    else:
        catalog["assets"].append({"id": asset_id, "source": source, "profile": "card"})
    catalog["assets"].sort(key=lambda entry: entry["id"])
    temporary_catalog = config.CATALOG_PATH.with_suffix(".tmp")
    temporary_catalog.write_text(json.dumps(catalog, indent=2) + "\n")
    temporary_catalog.replace(config.CATALOG_PATH)
    subprocess.run(["bun", "run", "assets"], cwd=config.PROJECT_ROOT / "frontend", check=True)
    if previous_source and previous_source != source:
        (config.PROJECT_ROOT / "assets" / "original" / previous_source).unlink()
