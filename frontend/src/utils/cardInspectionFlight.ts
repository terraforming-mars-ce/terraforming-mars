// The flight lives in the hand's stacking context, so neighboring cards naturally occlude it.
const REGIONS = [
  ".game-card-artwork",
  ".game-card-title",
  ".game-card-cost",
  ".game-card-tags",
  ".game-card-labels",
  ".game-card-behaviors",
  ".game-card-requirements",
];
const DURATION = 360;
const EASING = "cubic-bezier(0.22, 0.8, 0.3, 1)";

export interface CardInspectionFlight {
  play: (closing: boolean, speed?: number) => void;
  fadeBack: () => void;
  dispose: () => void;
}

function snapshot(card: HTMLElement, width: number, height: number) {
  const clone = card.cloneNode(true) as HTMLElement;
  // A snapshot has no React onLoad handler to reveal newly mounted artwork.
  clone.querySelectorAll<HTMLElement>(".game-card-artwork img").forEach((image) => {
    image.style.opacity = "1";
  });
  const stage = document.createElement("div");
  Object.assign(stage.style, {
    position: "fixed",
    left: "0",
    top: "0",
    opacity: "0",
    pointerEvents: "none",
  });
  Object.assign(clone.style, {
    width: `${card.offsetWidth}px`,
    fontFamily: getComputedStyle(card).fontFamily,
  });
  stage.inert = true;
  stage.append(clone);
  document.body.append(stage);
  const originalNodes = [card, ...card.querySelectorAll<HTMLElement>("*")];
  [clone, ...clone.querySelectorAll<HTMLElement>("*")].forEach((element, index) => {
    element.scrollTop = originalNodes[index].scrollTop;
    element.scrollLeft = originalNodes[index].scrollLeft;
  });
  const body = clone.querySelector<HTMLElement>(".game-card-body")!;
  const bodyRect = body.getBoundingClientRect();
  const cardRect = clone.getBoundingClientRect();
  const scaleX = width / bodyRect.width;
  const scaleY = height / bodyRect.height;
  const regions = REGIONS.flatMap((selector) => {
    const element = clone.querySelector<HTMLElement>(selector);
    if (!element) {
      return [];
    }
    const rect = element.getBoundingClientRect();
    return [
      {
        selector,
        element,
        x: (rect.x - bodyRect.x) * scaleX,
        y: (rect.y - bodyRect.y) * scaleY,
        width: rect.width * scaleX,
        height: rect.height * scaleY,
      },
    ];
  });
  const face = document.createElement("div");
  Object.assign(face.style, {
    position: "absolute",
    left: "0",
    top: "0",
    width: `${bodyRect.width}px`,
    transformOrigin: "0 0",
    transform: `scale(${scaleX}, ${scaleY})`,
  });
  Object.assign(clone.style, {
    position: "absolute",
    left: "0",
    top: `${cardRect.y - bodyRect.y}px`,
  });
  face.append(clone);
  stage.remove();
  return { face, regions, scaleX, scaleY };
}

export function createCardInspectionFlight(
  source: HTMLElement,
  detail: HTMLElement,
  onFinish: (closing: boolean) => void,
): CardInspectionFlight | null {
  const compact = source.querySelector<HTMLElement>(":scope > .game-card");
  const compactBody = compact?.querySelector<HTMLElement>(".game-card-body");
  const detailBody = detail.querySelector<HTMLElement>(".game-card-body");
  if (!compact || !compactBody || !detailBody || !source.isConnected) {
    return null;
  }
  const width = compactBody.offsetWidth;
  const height = compactBody.offsetHeight;
  const compactRect = compactBody.getBoundingClientRect();
  const detailRect = detailBody.getBoundingClientRect();
  const small = snapshot(compact, width, height);
  const large = snapshot(detail, width, height);
  const flight = document.createElement("div");
  flight.className = "card-inspection-flight";
  flight.inert = true;
  flight.setAttribute("aria-hidden", "true");
  Object.assign(flight.style, {
    position: "absolute",
    left: `${compactBody.offsetLeft}px`,
    top: `${compactBody.offsetTop}px`,
    width: `${width}px`,
    height: `${height}px`,
    visibility: "visible",
    pointerEvents: "none",
    transformOrigin: "center",
    willChange: "transform",
  });
  flight.append(small.face, large.face);
  source.append(flight);
  source.dataset.cardFlight = "true";

  let linear = new DOMMatrix();
  for (let element: HTMLElement | null = source; element; element = element.parentElement) {
    const transform = new DOMMatrix(getComputedStyle(element).transform);
    linear = new DOMMatrix([transform.a, transform.b, transform.c, transform.d, 0, 0]).multiply(
      linear,
    );
  }
  const delta = linear
    .inverse()
    .transformPoint(
      new DOMPoint(
        detailRect.x + detailRect.width / 2 - compactRect.x - compactRect.width / 2,
        detailRect.y + detailRect.height / 2 - compactRect.y - compactRect.height / 2,
      ),
    );
  const angle = (Math.atan2(linear.b, linear.a) * 180) / Math.PI;
  const scaleX = detailRect.width / width / Math.hypot(linear.a, linear.b);
  const scaleY = detailRect.height / height / Math.hypot(linear.c, linear.d);
  const expanded = `translate(${delta.x}px, ${delta.y}px) rotate(${-angle}deg) scale(${scaleX}, ${scaleY})`;
  const animations: Animation[] = [];
  const animate = (element: Element, frames: Keyframe[]) => {
    const animation = element.animate(frames, { duration: DURATION, easing: EASING, fill: "both" });
    animation.pause();
    animations.push(animation);
    return animation;
  };
  const movement = animate(flight, [{ transform: "none" }, { transform: expanded }]);
  animate(small.face, [
    { opacity: 1, offset: 0 },
    { opacity: 1, offset: 0.25 },
    { opacity: 0, offset: 0.85 },
    { opacity: 0 },
  ]);
  animate(large.face, [
    { opacity: 0, offset: 0 },
    { opacity: 0, offset: 0.25 },
    { opacity: 1, offset: 0.85 },
    { opacity: 1 },
  ]);
  for (const [face, other, compactFace] of [
    [small, large, true],
    [large, small, false],
  ] as const) {
    for (const region of face.regions) {
      const target = other.regions.find((entry) => entry.selector === region.selector);
      if (!target || !region.width || !region.height || !target.width || !target.height) {
        continue;
      }
      const transform = `translate(${(target.x - region.x) / face.scaleX}px, ${(target.y - region.y) / face.scaleY}px) scale(${target.width / region.width}, ${target.height / region.height})`;
      region.element.style.transformOrigin = "0 0";
      animate(
        region.element,
        compactFace
          ? [{ transform: "none" }, { transform }]
          : [{ transform }, { transform: "none" }],
      );
    }
  }
  let closing = false;
  let started = false;
  let fade: Animation | null = null;
  const cancelFade = () => {
    if (fade) {
      fade.onfinish = null;
      fade.cancel();
      fade = null;
    }
  };
  return {
    play(nextClosing, speed = 1) {
      const opacity = getComputedStyle(flight).opacity;
      cancelFade();
      if (Number(opacity) < 1) {
        fade = flight.animate([{ opacity }, { opacity: 1 }], {
          duration: 140,
          fill: "both",
        });
      }
      closing = nextClosing;
      movement.onfinish = () => onFinish(closing);
      const initialTime = closing ? DURATION : 0;
      const time = started ? Number(movement.currentTime) : initialTime;
      started = true;
      if (!closing && time >= DURATION) {
        if (fade) {
          fade.onfinish = () => onFinish(false);
        } else {
          movement.finish();
        }
        return;
      }
      for (const animation of animations) {
        animation.currentTime = time;
        animation.playbackRate = closing ? -speed : 1.25 * speed;
        animation.play();
      }
    },
    fadeBack() {
      const opacity = getComputedStyle(flight).opacity;
      cancelFade();
      movement.onfinish = null;
      const time = started ? Number(movement.currentTime) : DURATION;
      started = true;
      for (const animation of animations) {
        animation.pause();
        animation.currentTime = time;
      }
      fade = flight.animate([{ opacity }, { opacity: 0 }], {
        duration: 120,
        easing: "ease-out",
        fill: "both",
      });
      fade.onfinish = () => {
        for (const animation of animations) {
          animation.currentTime = 0;
        }
        cancelFade();
        fade = flight.animate([{ opacity: 0 }, { opacity: 1 }], {
          duration: 140,
          easing: "ease-in",
          fill: "both",
        });
        fade.onfinish = () => onFinish(true);
      };
    },
    dispose() {
      cancelFade();
      movement.onfinish = null;
      animations.forEach((animation) => animation.cancel());
      flight.remove();
      delete source.dataset.cardFlight;
    },
  };
}
