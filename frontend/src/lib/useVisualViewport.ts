import { useEffect, useState } from "react";

// キーボードとみなす見える範囲の縮み。アドレスバーの伸び縮みでは切り替えない
const KEYBOARD_THRESHOLD = 120;

// キーボードを除いた見える範囲の高さ。iOS は interactive-widget に対応せずページごとずらすため、枠の高さを合わせてずれを戻す
export const useVisualViewport = () => {
  const [state, setState] = useState<{ height: number; keyboardOpen: boolean }>();

  useEffect(() => {
    const viewport = globalThis.visualViewport;
    if (!viewport) {
      return undefined;
    }
    const update = () => {
      setState({
        height: viewport.height,
        keyboardOpen: globalThis.innerHeight - viewport.height > KEYBOARD_THRESHOLD,
      });
      globalThis.scrollTo(0, 0);
    };
    update();
    viewport.addEventListener("resize", update);
    return () => {
      viewport.removeEventListener("resize", update);
    };
  }, []);

  return state;
};
