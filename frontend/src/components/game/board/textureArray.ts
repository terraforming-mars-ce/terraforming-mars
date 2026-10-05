import * as THREE from "three";

const cache = new WeakMap<THREE.WebGLRenderer, Map<string, THREE.DataArrayTexture>>();

// Packs same-sized 2D textures into one sampler2DArray so materials stay within the 16 fragment
// texture units. Layers are uploaded straight from the decoded images (no canvas round-trip, so
// the height data in alpha is not premultiplied) and mipmapped on the GPU. Arrays are uploaded
// without flipY, so shaders sample them with a negated v to match the original 2D textures.
export function textureArray(
  gl: THREE.WebGLRenderer,
  textures: THREE.Texture[],
  colorSpace: THREE.ColorSpace,
): THREE.DataArrayTexture {
  const key = colorSpace + ":" + textures.map((texture) => texture.uuid).join(",");
  let arrays = cache.get(gl);
  if (!arrays) {
    arrays = new Map();
    cache.set(gl, arrays);
  }
  const cached = arrays.get(key);
  if (cached) {
    return cached;
  }
  const { width, height } = textures[0].image as { width: number; height: number };
  for (const texture of textures) {
    const image = texture.image as { width: number; height: number };
    if (image.width !== width || image.height !== height) {
      throw new Error(`Texture array layers must be ${width}x${height}`);
    }
  }
  const array = new THREE.DataArrayTexture(null, width, height, textures.length);
  // Allocate storage only; the layers are copied in below.
  array.source.dataReady = false;
  array.colorSpace = colorSpace;
  array.wrapS = array.wrapT = THREE.RepeatWrapping;
  array.minFilter = THREE.LinearMipmapLinearFilter;
  array.magFilter = THREE.LinearFilter;
  array.generateMipmaps = true;
  array.anisotropy = 4;
  array.needsUpdate = true;
  gl.initTexture(array);
  textures.forEach((texture, layer) => {
    gl.copyTextureToTexture(texture, array, null, new THREE.Vector3(0, 0, layer));
  });
  arrays.set(key, array);
  return array;
}
