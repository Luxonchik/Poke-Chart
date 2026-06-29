"use client";

import { useRef, useCallback } from "react";

const FOIL_GRADIENT =
  "repeating-linear-gradient(115deg, hsla(0,0%,100%,0) 0%, hsla(0,90%,65%,.5) 8%, hsla(48,95%,60%,.5) 16%, hsla(140,90%,60%,.5) 26%, hsla(190,95%,62%,.5) 36%, hsla(280,90%,66%,.5) 46%, hsla(0,0%,100%,0) 66%)";

const TILT_DEG = 14;

export function useCardTilt() {
  const cardRef = useRef<HTMLDivElement>(null);
  const foilRef = useRef<HTMLDivElement>(null);
  const glareRef = useRef<HTMLDivElement>(null);

  const onMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    const el = cardRef.current;
    const foil = foilRef.current;
    const glare = glareRef.current;
    if (!el || !foil || !glare) return;

    const rect = el.getBoundingClientRect();
    const px = (e.clientX - rect.left) / rect.width;
    const py = (e.clientY - rect.top) / rect.height;

    el.style.transform =
      `rotateX(${(py - 0.5) * -TILT_DEG}deg) rotateY(${(px - 0.5) * TILT_DEG}deg)`;

    foil.style.backgroundImage = FOIL_GRADIENT;
    foil.style.backgroundSize = "230% 230%";
    foil.style.backgroundPosition = `${px * 100}% ${py * 100}%`;
    foil.style.mixBlendMode = "color-dodge";
    foil.style.opacity = "1";

    glare.style.opacity = "1";
    glare.style.background =
      `radial-gradient(circle at ${px * 100}% ${py * 100}%, rgba(255,255,255,.5), transparent 42%)`;
  }, []);

  const onMouseLeave = useCallback(() => {
    const el = cardRef.current;
    const foil = foilRef.current;
    const glare = glareRef.current;
    if (!el || !foil || !glare) return;

    el.style.transform = "";
    foil.style.opacity = "0";
    glare.style.opacity = "0";
  }, []);

  return { cardRef, foilRef, glareRef, onMouseMove, onMouseLeave };
}
