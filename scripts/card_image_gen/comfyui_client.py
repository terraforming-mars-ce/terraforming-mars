"""ComfyUI API client for Flux Schnell image generation."""

import asyncio
import json
import uuid
from typing import Any

import aiohttp

from . import config


def build_flux_workflow(prompt: str, seed: int) -> dict[str, Any]:
    """Build a ComfyUI workflow for Flux Schnell generation.

    Node graph:
      1: UNETLoader         → model
      2: DualCLIPLoader      → clip
      3: VAELoader           → vae
      4: CLIPTextEncodeFlux  → positive conditioning
      5: EmptyLatentImage    → latent
      6: KSampler            → sampled latent
      7: VAEDecode           → image
      8: SaveImage           → disk
    """
    return {
        "1": {
            "class_type": "UNETLoader",
            "inputs": {
                "unet_name": config.UNET_MODEL,
                "weight_dtype": "default",
            },
        },
        "2": {
            "class_type": "DualCLIPLoader",
            "inputs": {
                "clip_name1": config.CLIP_L_MODEL,
                "clip_name2": config.T5XXL_MODEL,
                "type": "flux",
            },
        },
        "3": {
            "class_type": "VAELoader",
            "inputs": {
                "vae_name": config.VAE_MODEL,
            },
        },
        "4": {
            "class_type": "CLIPTextEncodeFlux",
            "inputs": {
                "clip": ["2", 0],
                "clip_l": prompt,
                "t5xxl": prompt,
                "guidance": config.GUIDANCE,
            },
        },
        "5": {
            "class_type": "EmptyLatentImage",
            "inputs": {
                "width": config.IMAGE_WIDTH,
                "height": config.IMAGE_HEIGHT,
                "batch_size": 1,
            },
        },
        "6": {
            "class_type": "KSampler",
            "inputs": {
                "model": ["1", 0],
                "positive": ["4", 0],
                "negative": ["4", 0],
                "latent_image": ["5", 0],
                "seed": seed,
                "steps": config.STEPS,
                "cfg": config.CFG,
                "sampler_name": config.SAMPLER,
                "scheduler": config.SCHEDULER,
                "denoise": 1.0,
            },
        },
        "7": {
            "class_type": "VAEDecode",
            "inputs": {
                "samples": ["6", 0],
                "vae": ["3", 0],
            },
        },
        "8": {
            "class_type": "SaveImage",
            "inputs": {
                "images": ["7", 0],
                "filename_prefix": "tfm_card",
            },
        },
    }


class ComfyUIClient:
    """Async client for the ComfyUI HTTP API."""

    def __init__(self, base_url: str = config.COMFYUI_URL):
        self.base_url = base_url
        self.client_id = str(uuid.uuid4())

    async def queue_prompt(self, workflow: dict[str, Any]) -> str:
        """Submit a workflow and return the prompt_id."""
        payload = {"prompt": workflow, "client_id": self.client_id}
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/prompt", json=payload
            ) as resp:
                resp.raise_for_status()
                data = await resp.json()
                return data["prompt_id"]

    async def wait_for_completion(
        self, prompt_id: str, poll_interval: float = 0.5, timeout: float = 300.0
    ) -> dict[str, Any]:
        """Poll the history endpoint until the prompt finishes."""
        elapsed = 0.0
        async with aiohttp.ClientSession() as session:
            while elapsed < timeout:
                async with session.get(
                    f"{self.base_url}/history/{prompt_id}"
                ) as resp:
                    history = await resp.json()
                    if prompt_id in history:
                        return history[prompt_id]
                await asyncio.sleep(poll_interval)
                elapsed += poll_interval
        raise TimeoutError(f"Generation timed out after {timeout}s")

    async def get_image(self, filename: str, subfolder: str = "") -> bytes:
        """Download a generated image from ComfyUI's output."""
        params = {"filename": filename, "subfolder": subfolder, "type": "output"}
        async with aiohttp.ClientSession() as session:
            async with session.get(
                f"{self.base_url}/view", params=params
            ) as resp:
                resp.raise_for_status()
                return await resp.read()

    async def generate(self, prompt: str, seed: int) -> bytes:
        """Full pipeline: build workflow, queue, wait, retrieve image bytes."""
        workflow = build_flux_workflow(prompt, seed)
        prompt_id = await self.queue_prompt(workflow)
        result = await self.wait_for_completion(prompt_id)

        # Walk outputs to find the generated image
        outputs = result.get("outputs", {})
        for output in outputs.values():
            if "images" in output:
                img_info = output["images"][0]
                return await self.get_image(
                    img_info["filename"], img_info.get("subfolder", "")
                )

        raise RuntimeError("No image found in generation output")
