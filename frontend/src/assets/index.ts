import { assets, corporations } from "./generated/registry.ts";
import type { AssetId, AssetVariant } from "./generated/registry.ts";
import { serverUrl } from "../config.ts";
export { corporations };
export type { AssetId };

export function assetVariant(id: AssetId, pixels = Infinity): AssetVariant {
  const variants = assets[id].variants;
  const variant =
    variants.find((variant) => Math.max(variant.width, variant.height) >= pixels) ??
    variants[variants.length - 1];
  return { ...variant, url: serverUrl(variant.url) };
}
export function assetUrl(id: AssetId, pixels?: number): string {
  return assetVariant(id, pixels).url;
}
export function assetImage(id: AssetId, sizes: string, pixels?: number) {
  const variants = assets[id].variants;
  return {
    src: assetUrl(id, pixels),
    srcSet: variants.map((variant) => `${serverUrl(variant.url)} ${variant.width}w`).join(", "),
    sizes,
  };
}
export function cardImage(cardId: string) {
  const id = `cards/${cardId}`;
  if (!(id in assets)) {
    return undefined;
  }
  return assetImage(id as AssetId, "(max-width: 600px) 90vw, 400px", 480);
}
