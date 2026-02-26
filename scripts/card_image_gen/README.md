# Card Image Generation

Generates card artwork for Terraforming Mars using [ComfyUI](https://github.com/comfyanonymous/ComfyUI) with [Flux Schnell](https://huggingface.co/black-forest-labs/FLUX.1-schnell) (FP8 quantized).

## Prerequisites

- NVIDIA GPU with 8GB+ VRAM (tested on RTX 4060)
- Python 3.10+
- ~17GB disk space for models
- [uv](https://docs.astral.sh/uv/) for running the generation script (or install deps manually with pip)

## ComfyUI Setup

### 1. Clone ComfyUI

```bash
git clone https://github.com/comfyanonymous/ComfyUI.git
cd ComfyUI
```

### 2. Create a virtual environment and install dependencies

```bash
python -m venv venv
source venv/bin/activate    # Linux/macOS
# venv\Scripts\activate     # Windows

pip install --upgrade pip
pip install -r requirements.txt
```

### 3. Install PyTorch with CUDA

Pick the command matching your CUDA version from https://pytorch.org/get-started/locally/.

For CUDA 12.8:

```bash
pip install torch torchvision torchaudio --index-url https://download.pytorch.org/whl/cu128
```

### 4. Download models

From inside the ComfyUI directory, with the venv activated:

```bash
pip install huggingface_hub   # if not already installed

python -c "
from huggingface_hub import hf_hub_download

# Flux Schnell FP8 checkpoint (~12GB)
hf_hub_download(
    repo_id='Kijai/flux-fp8',
    filename='flux1-schnell-fp8-e4m3fn.safetensors',
    local_dir='models/checkpoints'
)

# Flux VAE (~160MB)
hf_hub_download(
    repo_id='Kijai/flux-fp8',
    filename='flux-vae-bf16.safetensors',
    local_dir='models/vae'
)

# CLIP-L text encoder (~235MB)
hf_hub_download(
    repo_id='comfyanonymous/flux_text_encoders',
    filename='clip_l.safetensors',
    local_dir='models/text_encoders'
)

# T5-XXL text encoder, FP8 (~4.6GB)
hf_hub_download(
    repo_id='comfyanonymous/flux_text_encoders',
    filename='t5xxl_fp8_e4m3fn.safetensors',
    local_dir='models/text_encoders'
)
"
```

### 5. Symlink the checkpoint for UNETLoader

The generation script uses `UNETLoader` which reads from `models/diffusion_models/`. Create a symlink so it can find the checkpoint:

```bash
# Linux/macOS
ln -sf ../checkpoints/flux1-schnell-fp8-e4m3fn.safetensors models/diffusion_models/flux1-schnell-fp8-e4m3fn.safetensors

# Windows (run as admin)
# mklink models\diffusion_models\flux1-schnell-fp8-e4m3fn.safetensors models\checkpoints\flux1-schnell-fp8-e4m3fn.safetensors
```

### 6. Start ComfyUI

For GPUs with 8GB VRAM, use `--lowvram` to enable automatic model offloading:

```bash
python main.py --lowvram
```

For GPUs with 12GB+ VRAM, you can omit the flag:

```bash
python main.py
```

Add `--listen 0.0.0.0` if you want to access the web UI from another device on the network.

ComfyUI will be available at http://localhost:8188.

## Configuration

Before running, update `config.py` to match your setup:

- `COMFYUI_URL` - ComfyUI server address (default: `http://127.0.0.1:8188`)
- `COMFYUI_PATH` - Path to your ComfyUI installation
- `IMAGE_WIDTH` / `IMAGE_HEIGHT` - Output dimensions (default: 960x720)
- `WEBP_QUALITY` - WebP compression quality (default: 90)
- `STEPS` - Sampling steps (default: 4; Flux Schnell is optimized for 4)
- `GUIDANCE` - Guidance scale (default: 3.5)

## Generating Card Images

Make sure ComfyUI is running, then run from the project root:

```bash
# Preview the prompt for a card without generating
uv run --with aiohttp --with Pillow \
  python -m scripts.card_image_gen.generate_cards --card 042 --dry-run

# Generate a single card
uv run --with aiohttp --with Pillow \
  python -m scripts.card_image_gen.generate_cards --card 042

# Generate a range of cards
uv run --with aiohttp --with Pillow \
  python -m scripts.card_image_gen.generate_cards --start 042 --end 100

# Generate all cards that don't have images yet
uv run --with aiohttp --with Pillow \
  python -m scripts.card_image_gen.generate_cards --missing

# Use a fixed seed for reproducibility
uv run --with aiohttp --with Pillow \
  python -m scripts.card_image_gen.generate_cards --card 042 --seed 12345
```

If you don't use `uv`, install the dependencies from `requirements.txt` into a venv and call the module directly:

```bash
pip install -r scripts/card_image_gen/requirements.txt
python -m scripts.card_image_gen.generate_cards --card 042
```

Images are saved as 960x720 WebP files to `frontend/public/assets/cards/{id}.webp`.

## How It Works

1. Reads card data from `backend/assets/terraforming_mars_cards.json`
2. Builds a prompt from the card name, tags, and type using a consistent sci-fi Martian art style
3. Sends a Flux Schnell workflow to ComfyUI's API (UNETLoader + DualCLIPLoader + VAELoader + KSampler)
4. Retrieves the generated image, resizes to 960x720, and saves as WebP

## Troubleshooting

**`BrokenPipeError` during generation** - Your GPU is running out of VRAM. Restart ComfyUI with `--lowvram`.

**`Connection refused`** - ComfyUI is not running. Start it first (see step 6 above).

**`Entry Not Found` when downloading models** - The Hugging Face repo or filename may have changed. Check the repos manually:
- https://huggingface.co/Kijai/flux-fp8
- https://huggingface.co/comfyanonymous/flux_text_encoders
