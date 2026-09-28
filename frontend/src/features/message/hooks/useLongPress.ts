import { useEffect, useRef, useState } from "react";
import type { HTMLAttributes, PointerEvent } from "react";

const LONG_PRESS_MS = 450;
// これ以上指が動いたらスクロールとみなして取り消す
const MOVE_TOLERANCE_PX = 8;

// 長押し（と右クリック）で onLongPress を呼ぶ。isEnabled が false の間は何もしない
export const useLongPress = (onLongPress: () => void, isEnabled: boolean) => {
  const timerRef = useRef<ReturnType<typeof setTimeout>>(undefined);
  const originRef = useRef({ x: 0, y: 0 });
  // 長押しの直後に指を離したときのクリックでリンクなどが反応しないようにする
  const firedRef = useRef(false);
  const [isPressed, setIsPressed] = useState(false);

  useEffect(
    () => () => {
      clearTimeout(timerRef.current);
    },
    [],
  );

  const cancel = () => {
    clearTimeout(timerRef.current);
    setIsPressed(false);
  };

  const fire = () => {
    cancel();
    firedRef.current = true;
    onLongPress();
  };

  if (!isEnabled) {
    return { isPressed: false, longPressProps: {} };
  }

  const longPressProps: HTMLAttributes<HTMLElement> = {
    onClickCapture: (event) => {
      if (firedRef.current) {
        firedRef.current = false;
        event.preventDefault();
        event.stopPropagation();
      }
    },
    onContextMenu: (event) => {
      event.preventDefault();
      fire();
    },
    onPointerCancel: cancel,
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      if (event.button !== 0) {
        return;
      }
      firedRef.current = false;
      originRef.current = { x: event.clientX, y: event.clientY };
      setIsPressed(true);
      timerRef.current = setTimeout(fire, LONG_PRESS_MS);
    },
    onPointerLeave: cancel,
    onPointerMove: (event: PointerEvent<HTMLElement>) => {
      const { x, y } = originRef.current;
      if (Math.hypot(event.clientX - x, event.clientY - y) > MOVE_TOLERANCE_PX) {
        cancel();
      }
    },
    onPointerUp: cancel,
  };

  return { isPressed, longPressProps };
};
