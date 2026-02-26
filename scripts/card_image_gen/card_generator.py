"""Main orchestration for card image generation."""

import asyncio
import json
import random
from pathlib import Path
from typing import Optional

from . import config
from .comfyui_client import ComfyUIClient
from .image_processor import process_and_save
from .prompt_builder import build_prompt


async def _generate_with_retry(
    client: ComfyUIClient, prompt: str, seed: int
) -> bytes:
    """Generate an image with retry logic."""
    last_error: Optional[Exception] = None
    for attempt in range(config.MAX_RETRIES):
        try:
            return await client.generate(prompt, seed)
        except Exception as exc:
            last_error = exc
            if attempt < config.MAX_RETRIES - 1:
                delay = config.RETRY_DELAY * (attempt + 1)
                print(f"  Retry {attempt + 1}/{config.MAX_RETRIES} after error: {exc}")
                await asyncio.sleep(delay)
    raise last_error  # type: ignore[misc]


class CardImageGenerator:
    """Generates card images using ComfyUI with Flux Schnell."""

    def __init__(self) -> None:
        self.client = ComfyUIClient()
        self.cards = self._load_cards()
        self.output_dir = Path(config.OUTPUT_DIR)

    @staticmethod
    def _load_cards() -> list[dict]:
        with open(config.CARD_JSON_PATH) as f:
            return json.load(f)

    def find_card(self, card_id: str) -> dict:
        """Look up a card by its ID."""
        for card in self.cards:
            if card["id"] == card_id:
                return card
        raise ValueError(f"Card '{card_id}' not found")

    def get_missing_cards(self) -> list[dict]:
        """Return cards that don't have generated images yet."""
        return [
            card
            for card in self.cards
            if not (self.output_dir / f"{card['id']}.webp").exists()
        ]

    async def generate_card(
        self, card: dict, seed: Optional[int] = None
    ) -> Path:
        """Generate an image for a single card and save it."""
        if seed is None:
            seed = random.randint(0, 2**32 - 1)

        prompt = build_prompt(card)
        print(f"[{card['id']}] {card['name']}")
        print(f"  Prompt: {prompt[:120]}...")
        print(f"  Seed:   {seed}")

        image_data = await _generate_with_retry(self.client, prompt, seed)

        output_path = self.output_dir / f"{card['id']}.webp"
        process_and_save(image_data, output_path)

        print(f"  Saved:  {output_path}")
        return output_path

    async def generate_batch(
        self, cards: list[dict], seed: Optional[int] = None
    ) -> list[Path]:
        """Generate images for multiple cards sequentially."""
        results: list[Path] = []
        total = len(cards)
        for i, card in enumerate(cards, 1):
            print(f"\n--- Card {i}/{total} ---")
            try:
                path = await self.generate_card(card, seed=seed)
                results.append(path)
            except Exception as exc:
                print(f"  FAILED: {exc}")

            if i < total:
                await asyncio.sleep(config.BATCH_DELAY)

        print(f"\nDone. Generated {len(results)}/{total} images.")
        return results
