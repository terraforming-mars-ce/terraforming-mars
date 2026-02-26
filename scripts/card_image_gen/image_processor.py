"""Post-processing for generated card images."""

from io import BytesIO
from pathlib import Path

from PIL import Image

from . import config


def process_and_save(
    image_data: bytes,
    output_path: Path,
    width: int = config.IMAGE_WIDTH,
    height: int = config.IMAGE_HEIGHT,
    quality: int = config.WEBP_QUALITY,
) -> None:
    """Resize to exact dimensions if needed and save as WebP."""
    img = Image.open(BytesIO(image_data))

    if img.size != (width, height):
        img = img.resize((width, height), Image.Resampling.LANCZOS)

    if img.mode != "RGB":
        img = img.convert("RGB")

    img.save(output_path, "WEBP", quality=quality)
