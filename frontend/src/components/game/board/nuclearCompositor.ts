import * as THREE from "three";
import type { NuclearEffect } from "../../../contexts/NuclearEffectsContext";
import { nuclearShake } from "./nuclearGeometry";
import vertexShader from "./shaders/nuclear-screen.vert.glsl?raw";
import cloudFragment from "./shaders/nuclear-cloud.frag.glsl?raw";
import compositeFragment from "./shaders/nuclear-composite.frag.glsl?raw";

export class NuclearCompositor {
  private depthTarget: THREE.WebGLRenderTarget | null = null;
  private cloudTarget: THREE.WebGLRenderTarget | null = null;
  private geometry = new THREE.PlaneGeometry(2, 2);
  private screenCamera = new THREE.Camera();
  private volumeScene = new THREE.Scene();
  private compositeScene = new THREE.Scene();
  private cloud = new THREE.ShaderMaterial({
    vertexShader,
    fragmentShader: cloudFragment,
    uniforms: {
      uDepth: { value: null },
      uInverseProjection: { value: new THREE.Matrix4() },
      uCameraWorld: { value: new THREE.Matrix4() },
      uWorldToLocal: { value: new THREE.Matrix4() },
      uAge: { value: -1 },
      uSeed: { value: 1 },
    },
    transparent: true,
    premultipliedAlpha: true,
    depthTest: false,
    depthWrite: false,
    toneMapped: false,
  });
  private composite = new THREE.ShaderMaterial({
    vertexShader,
    fragmentShader: compositeFragment,
    uniforms: {
      uCloud: { value: null },
      uDepth: { value: null },
      uCloudTexel: { value: new THREE.Vector2(1, 1) },
      uInverseProjection: { value: new THREE.Matrix4() },
    },
    depthTest: false,
    depthWrite: false,
    transparent: true,
    premultipliedAlpha: true,
  });
  private size = new THREE.Vector2();
  private clearColor = new THREE.Color();
  private previousScissor = new THREE.Vector4();
  private projection = new THREE.Matrix4();
  private projectionInverse = new THREE.Matrix4();
  private frustum = new THREE.Frustum();
  private viewProjection = new THREE.Matrix4();
  private sphere = new THREE.Sphere();
  private position = new THREE.Vector3();
  private up = new THREE.Vector3();
  private toCamera = new THREE.Vector3();
  private corner = new THREE.Vector3();
  private warmed = false;

  constructor() {
    const volume = new THREE.Mesh(this.geometry, this.cloud);
    volume.frustumCulled = false;
    this.volumeScene.add(volume);
    const composite = new THREE.Mesh(this.geometry, this.composite);
    composite.frustumCulled = false;
    this.compositeScene.add(composite);
  }

  private targets(gl: THREE.WebGLRenderer, width: number, height: number) {
    const scale = Math.min(0.5, 960 / width);
    const lowWidth = Math.max(1, Math.round(width * scale));
    const lowHeight = Math.max(1, Math.round(height * scale));
    if (!this.depthTarget) {
      const type = gl.extensions.has("EXT_color_buffer_float")
        ? THREE.HalfFloatType
        : THREE.UnsignedByteType;
      const context = gl.getContext() as WebGL2RenderingContext;
      const depthType =
        context.getParameter(context.DEPTH_BITS) === 32
          ? THREE.FloatType
          : THREE.UnsignedInt248Type;
      const depth = new THREE.DepthTexture(width, height, depthType);
      depth.format = THREE.DepthStencilFormat;
      depth.minFilter = depth.magFilter = THREE.NearestFilter;
      this.depthTarget = new THREE.WebGLRenderTarget(width, height, {
        depthTexture: depth,
        depthBuffer: true,
        stencilBuffer: true,
      });
      this.cloudTarget = new THREE.WebGLRenderTarget(lowWidth, lowHeight, {
        type,
        depthBuffer: false,
        minFilter: THREE.LinearFilter,
        magFilter: THREE.LinearFilter,
      });
    }
    this.depthTarget.setSize(width, height);
    this.cloudTarget!.setSize(lowWidth, lowHeight);
    this.cloud.uniforms.uDepth.value = this.depthTarget.depthTexture;
    this.composite.uniforms.uDepth.value = this.depthTarget.depthTexture;
    this.composite.uniforms.uCloud.value = this.cloudTarget!.texture;
    this.composite.uniforms.uCloudTexel.value.set(1 / lowWidth, 1 / lowHeight);
  }

  private releaseTargets() {
    this.depthTarget?.dispose();
    this.cloudTarget?.dispose();
    this.depthTarget = null;
    this.cloudTarget = null;
    this.cloud.uniforms.uDepth.value = null;
    this.composite.uniforms.uDepth.value = null;
    this.composite.uniforms.uCloud.value = null;
  }

  private scissor(gl: THREE.WebGLRenderer, effect: NuclearEffect, camera: THREE.Camera) {
    let minX = 1,
      minY = 1,
      maxX = -1,
      maxY = -1;
    for (let i = 0; i < 8; i++) {
      this.corner
        .set(i & 1 ? 0.36 : -0.36, i & 2 ? 0.36 : -0.36, i & 4 ? 0.5 : -0.06)
        .applyMatrix4(effect.object.matrixWorld)
        .project(camera);
      if (this.corner.z < -1 || this.corner.z > 1) {
        minX = minY = -1;
        maxX = maxY = 1;
        break;
      }
      minX = Math.min(minX, this.corner.x);
      maxX = Math.max(maxX, this.corner.x);
      minY = Math.min(minY, this.corner.y);
      maxY = Math.max(maxY, this.corner.y);
    }
    const width = this.cloudTarget!.width,
      height = this.cloudTarget!.height;
    const x = Math.max(0, Math.floor((minX + 1) * 0.5 * width));
    const y = Math.max(0, Math.floor((minY + 1) * 0.5 * height));
    const dpr = gl.getPixelRatio();
    gl.setScissor(
      x / dpr,
      y / dpr,
      Math.max(0, Math.min(width, Math.ceil((maxX + 1) * 0.5 * width)) - x) / dpr,
      Math.max(0, Math.min(height, Math.ceil((maxY + 1) * 0.5 * height)) - y) / dpr,
    );
  }

  render(
    gl: THREE.WebGLRenderer,
    scene: THREE.Scene,
    camera: THREE.Camera,
    effects: Set<NuclearEffect>,
    shakeEnabled: boolean,
    cssWidth: number,
    cssHeight: number,
  ) {
    if (this.warmed && effects.size === 0) {
      if (this.depthTarget) {
        this.releaseTargets();
      }
      gl.render(scene, camera);
      return;
    }
    const oldAutoClear = gl.autoClear,
      oldTarget = gl.getRenderTarget();
    const oldAlpha = gl.getClearAlpha();
    const oldScissorTest = gl.getScissorTest();
    gl.getScissor(this.previousScissor);
    gl.getClearColor(this.clearColor);
    this.projection.copy(camera.projectionMatrix);
    this.projectionInverse.copy(camera.projectionMatrixInverse);
    try {
      if (!this.warmed) {
        this.targets(gl, 1, 1);
        gl.setRenderTarget(this.cloudTarget);
        gl.render(this.volumeScene, this.screenCamera);
        gl.setRenderTarget(oldTarget);
        gl.compile(this.compositeScene, this.screenCamera);
        this.releaseTargets();
        this.warmed = true;
      }
      this.viewProjection.multiplyMatrices(camera.projectionMatrix, camera.matrixWorldInverse);
      this.frustum.setFromProjectionMatrix(this.viewProjection);
      const visible: NuclearEffect[] = [];
      let shakeX = 0,
        shakeY = 0;
      for (const effect of effects) {
        if (effect.age < 0 || effect.age >= 6.2) {
          continue;
        }
        this.sphere.center.set(0, 0, 0.2).applyMatrix4(effect.object.matrixWorld);
        this.sphere.radius = 0.5;
        if (!this.frustum.intersectsSphere(this.sphere)) {
          continue;
        }
        this.position.setFromMatrixPosition(effect.object.matrixWorld);
        this.up.set(0, 0, 1).transformDirection(effect.object.matrixWorld);
        this.toCamera.copy(camera.position).sub(this.position);
        if (this.up.dot(this.toCamera) <= 0) {
          continue;
        }
        visible.push(effect);
        if (shakeEnabled && this.frustum.containsPoint(this.position)) {
          const shake = nuclearShake(effect.age);
          shakeX += shake[0];
          shakeY += shake[1];
        }
      }
      const magnitude = Math.hypot(shakeX, shakeY);
      const clamp = magnitude > 1.5 ? 1.5 / magnitude : 1;
      camera.projectionMatrix.elements[8] += (shakeX * clamp * 2) / cssWidth;
      camera.projectionMatrix.elements[9] += (shakeY * clamp * 2) / cssHeight;
      camera.projectionMatrixInverse.copy(camera.projectionMatrix).invert();
      if (visible.length === 0) {
        this.releaseTargets();
        gl.setRenderTarget(oldTarget);
        gl.render(scene, camera);
        return;
      }
      gl.getDrawingBufferSize(this.size);
      this.targets(gl, this.size.x, this.size.y);
      this.cloud.uniforms.uInverseProjection.value.copy(camera.projectionMatrixInverse);
      this.cloud.uniforms.uCameraWorld.value.copy(camera.matrixWorld);
      this.composite.uniforms.uInverseProjection.value.copy(camera.projectionMatrixInverse);
      // Render once to the usual framebuffer, preserving all existing shader color handling.
      gl.setRenderTarget(oldTarget);
      gl.render(scene, camera);
      const context = gl.getContext() as WebGL2RenderingContext;
      const sourceFramebuffer = context.getParameter(context.FRAMEBUFFER_BINDING);
      gl.setRenderTarget(this.depthTarget);
      context.bindFramebuffer(context.READ_FRAMEBUFFER, sourceFramebuffer);
      context.blitFramebuffer(
        0,
        0,
        this.size.x,
        this.size.y,
        0,
        0,
        this.size.x,
        this.size.y,
        context.DEPTH_BUFFER_BIT | context.STENCIL_BUFFER_BIT,
        context.NEAREST,
      );
      gl.autoClear = false;
      gl.setRenderTarget(this.cloudTarget);
      gl.setClearColor(0, 0);
      gl.clear(true, false, false);
      visible.sort((a, b) => {
        const da = this.position
          .setFromMatrixPosition(a.object.matrixWorld)
          .distanceToSquared(camera.position);
        const db = this.position
          .setFromMatrixPosition(b.object.matrixWorld)
          .distanceToSquared(camera.position);
        return db - da;
      });
      gl.setScissorTest(true);
      for (const effect of visible) {
        this.scissor(gl, effect, camera);
        this.cloud.uniforms.uWorldToLocal.value.copy(effect.object.matrixWorld).invert();
        this.cloud.uniforms.uAge.value = effect.age;
        this.cloud.uniforms.uSeed.value = effect.seed;
        gl.render(this.volumeScene, this.screenCamera);
      }
      gl.setScissorTest(false);
      gl.setRenderTarget(oldTarget);
      gl.render(this.compositeScene, this.screenCamera);
    } finally {
      camera.projectionMatrix.copy(this.projection);
      camera.projectionMatrixInverse.copy(this.projectionInverse);
      gl.setRenderTarget(oldTarget);
      gl.setScissor(this.previousScissor);
      gl.setScissorTest(oldScissorTest);
      gl.setClearColor(this.clearColor, oldAlpha);
      gl.autoClear = oldAutoClear;
    }
  }

  dispose() {
    this.releaseTargets();
    this.geometry.dispose();
    this.cloud.dispose();
    this.composite.dispose();
    this.warmed = false;
  }
}
