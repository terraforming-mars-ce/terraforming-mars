import { useRef, useEffect } from "react";
import { useThree, useFrame } from "@react-three/fiber";
import * as THREE from "three";
import { useWorld3DSettings, StoredCameraState } from "../../../contexts/World3DSettingsContext";

export function FreeCamera() {
  const { camera, gl } = useThree();
  const { settings, cameraStateRef, pendingCameraTransformRef } = useWorld3DSettings();

  const moveSpeed = 0.01;
  const lookSpeed = 0.0004;

  const keys = useRef({
    forward: false,
    backward: false,
    left: false,
    right: false,
    up: false,
    down: false,
  });

  const euler = useRef(new THREE.Euler(0, 0, 0, "YXZ"));
  const isPointerLocked = useRef(false);

  useEffect(() => {
    if (!settings.freeCameraEnabled) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.code) {
        case "KeyW":
          keys.current.forward = true;
          break;
        case "KeyS":
          keys.current.backward = true;
          break;
        case "KeyA":
          keys.current.left = true;
          break;
        case "KeyD":
          keys.current.right = true;
          break;
        case "Space":
          keys.current.up = true;
          break;
        case "ShiftLeft":
        case "ShiftRight":
          keys.current.down = true;
          break;
      }
    };

    const handleKeyUp = (e: KeyboardEvent) => {
      switch (e.code) {
        case "KeyW":
          keys.current.forward = false;
          break;
        case "KeyS":
          keys.current.backward = false;
          break;
        case "KeyA":
          keys.current.left = false;
          break;
        case "KeyD":
          keys.current.right = false;
          break;
        case "Space":
          keys.current.up = false;
          break;
        case "ShiftLeft":
        case "ShiftRight":
          keys.current.down = false;
          break;
      }
    };

    const handleMouseMove = (e: MouseEvent) => {
      if (!isPointerLocked.current) return;

      euler.current.setFromQuaternion(camera.quaternion);
      euler.current.y -= e.movementX * lookSpeed;
      euler.current.x -= e.movementY * lookSpeed;
      euler.current.x = Math.max(-Math.PI / 2, Math.min(Math.PI / 2, euler.current.x));
      camera.quaternion.setFromEuler(euler.current);
    };

    const handleClick = () => {
      gl.domElement.requestPointerLock();
    };

    const handlePointerLockChange = () => {
      isPointerLocked.current = document.pointerLockElement === gl.domElement;
      gl.domElement.style.cursor = isPointerLocked.current ? "none" : "crosshair";
    };

    gl.domElement.style.cursor = "crosshair";

    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("keyup", handleKeyUp);
    document.addEventListener("mousemove", handleMouseMove);
    gl.domElement.addEventListener("click", handleClick);
    document.addEventListener("pointerlockchange", handlePointerLockChange);

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("keyup", handleKeyUp);
      document.removeEventListener("mousemove", handleMouseMove);
      gl.domElement.removeEventListener("click", handleClick);
      document.removeEventListener("pointerlockchange", handlePointerLockChange);

      if (document.pointerLockElement === gl.domElement) {
        document.exitPointerLock();
      }
      gl.domElement.style.cursor = "grab";
    };
  }, [camera, gl, settings.freeCameraEnabled]);

  useFrame(() => {
    if (!settings.freeCameraEnabled) return;

    const pending = pendingCameraTransformRef.current;
    if (pending) {
      if (pending.position) {
        camera.position.set(pending.position.x, pending.position.y, pending.position.z);
      }
      if (pending.rotation) {
        euler.current.set(pending.rotation.x, pending.rotation.y, pending.rotation.z, "YXZ");
        camera.quaternion.setFromEuler(euler.current);
      }
      pendingCameraTransformRef.current = null;
    }

    const direction = new THREE.Vector3();
    const right = new THREE.Vector3();

    camera.getWorldDirection(direction);
    right.crossVectors(direction, camera.up).normalize();

    if (keys.current.forward) camera.position.addScaledVector(direction, moveSpeed);
    if (keys.current.backward) camera.position.addScaledVector(direction, -moveSpeed);
    if (keys.current.left) camera.position.addScaledVector(right, -moveSpeed);
    if (keys.current.right) camera.position.addScaledVector(right, moveSpeed);
    if (keys.current.up) camera.position.y += moveSpeed;
    if (keys.current.down) camera.position.y -= moveSpeed;

    cameraStateRef.current = {
      position: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
      rotation: { x: camera.rotation.x, y: camera.rotation.y, z: camera.rotation.z },
    };
  });

  return null;
}

interface CameraFrustumHelperProps {
  storedState: StoredCameraState;
  fov: number;
  aspect: number;
}

export function CameraFrustumHelper({ storedState, fov, aspect }: CameraFrustumHelperProps) {
  const { settings } = useWorld3DSettings();
  const groupRef = useRef<THREE.Group>(null);

  if (!settings.freeCameraEnabled || !settings.showCameraFrustum) {
    return null;
  }

  const position = new THREE.Vector3(
    storedState.position.x,
    storedState.position.y,
    storedState.position.z,
  );

  const near = 0.1;
  const far = 5;

  const halfFovRad = (fov * Math.PI) / 360;
  const nearHeight = 2 * Math.tan(halfFovRad) * near;
  const nearWidth = nearHeight * aspect;
  const farHeight = 2 * Math.tan(halfFovRad) * far;
  const farWidth = farHeight * aspect;

  const nTL = new THREE.Vector3(-nearWidth / 2, nearHeight / 2, -near);
  const nTR = new THREE.Vector3(nearWidth / 2, nearHeight / 2, -near);
  const nBL = new THREE.Vector3(-nearWidth / 2, -nearHeight / 2, -near);
  const nBR = new THREE.Vector3(nearWidth / 2, -nearHeight / 2, -near);

  const fTL = new THREE.Vector3(-farWidth / 2, farHeight / 2, -far);
  const fTR = new THREE.Vector3(farWidth / 2, farHeight / 2, -far);
  const fBL = new THREE.Vector3(-farWidth / 2, -farHeight / 2, -far);
  const fBR = new THREE.Vector3(farWidth / 2, -farHeight / 2, -far);

  const linePoints = [
    nTL,
    nTR,
    nTR,
    nBR,
    nBR,
    nBL,
    nBL,
    nTL,
    fTL,
    fTR,
    fTR,
    fBR,
    fBR,
    fBL,
    fBL,
    fTL,
    nTL,
    fTL,
    nTR,
    fTR,
    nBL,
    fBL,
    nBR,
    fBR,
  ];

  const geometry = new THREE.BufferGeometry().setFromPoints(linePoints);

  const lookTarget = new THREE.Vector3(0, 0, 0);
  const quaternion = new THREE.Quaternion();
  const up = new THREE.Vector3(0, 1, 0);
  const matrix = new THREE.Matrix4().lookAt(position, lookTarget, up);
  quaternion.setFromRotationMatrix(matrix);

  return (
    <group ref={groupRef} position={position} quaternion={quaternion}>
      <lineSegments geometry={geometry}>
        <lineBasicMaterial color="#ffff00" linewidth={2} />
      </lineSegments>
      <mesh position={[0, 0, 0]}>
        <sphereGeometry args={[0.05, 8, 8]} />
        <meshBasicMaterial color="#ffff00" />
      </mesh>
    </group>
  );
}
