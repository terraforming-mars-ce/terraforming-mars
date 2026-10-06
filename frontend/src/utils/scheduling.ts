export function afterNextPaint(callback: () => void): () => void {
  let timeout: ReturnType<typeof setTimeout> | undefined;
  // rAF runs before the next paint; a task queued from it runs after that paint.
  const frame = requestAnimationFrame(() => {
    timeout = setTimeout(callback);
  });
  return () => {
    cancelAnimationFrame(frame);
    clearTimeout(timeout);
  };
}

export function whenIdle(callback: () => void): () => void {
  if (typeof requestIdleCallback === "function") {
    const handle = requestIdleCallback(callback, { timeout: 1000 });
    return () => cancelIdleCallback(handle);
  }
  const timeout = setTimeout(callback, 200);
  return () => clearTimeout(timeout);
}
