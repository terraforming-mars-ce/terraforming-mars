"""Configuration constants for card image generation."""

from pathlib import Path

# ComfyUI settings
COMFYUI_URL = "http://127.0.0.1:8188"
COMFYUI_PATH = Path("/mnt/HDD/ComfyUI")

# Model files (relative to ComfyUI models folder)
UNET_MODEL = "flux1-schnell-fp8-e4m3fn.safetensors"
VAE_MODEL = "flux-vae-bf16.safetensors"
CLIP_L_MODEL = "clip_l.safetensors"
T5XXL_MODEL = "t5xxl_fp8_e4m3fn.safetensors"

# Project paths
PROJECT_ROOT = Path("/home/mafs/Documents/Repositories/terraforming-mars")
CARD_JSON_PATH = PROJECT_ROOT / "backend" / "assets" / "terraforming_mars_cards.json"
OUTPUT_DIR = PROJECT_ROOT / "frontend" / "public" / "assets" / "cards"

# Image specifications
IMAGE_WIDTH = 960
IMAGE_HEIGHT = 720
WEBP_QUALITY = 90

# Flux Schnell generation parameters
STEPS = 4
CFG = 1.0
SAMPLER = "euler"
SCHEDULER = "simple"
GUIDANCE = 3.5

# Retry settings
MAX_RETRIES = 3
RETRY_DELAY = 2.0

# Batch settings
BATCH_DELAY = 1.0  # Delay between generations in seconds
