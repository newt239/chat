import { createContext, useContext, useEffect } from "react";

type MobileStack = {
  back: () => void;
  // 左へスワイプしたときの操作を登録する。画面の中身が呼ぶ
  setForward: (forward: (() => void) | null) => void;
};

// モバイルで積み重ねた画面の操作。積まれた画面の中でだけ値を持つ
export const MobileStackContext = createContext<MobileStack | null>(null);

export const useMobileForward = (forward: () => void) => {
  const stack = useContext(MobileStackContext);
  useEffect(() => {
    stack?.setForward(forward);
    return () => {
      stack?.setForward(null);
    };
  }, [stack, forward]);
};
