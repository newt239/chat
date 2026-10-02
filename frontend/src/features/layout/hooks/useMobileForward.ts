import { createContext, useContext, useEffect } from "react";
import type { RefObject } from "react";

// 積み重ねた画面で左へスワイプしたときの操作。画面の中身が登録する
export const MobileForwardContext = createContext<RefObject<(() => void) | null> | null>(null);

export const useMobileForward = (forward: () => void) => {
  const ref = useContext(MobileForwardContext);
  useEffect(() => {
    if (ref !== null) {
      ref.current = forward;
    }
    return () => {
      if (ref !== null) {
        ref.current = null;
      }
    };
  }, [ref, forward]);
};
