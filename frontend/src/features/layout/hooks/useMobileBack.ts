import { createContext, useContext } from "react";

// モバイルで積み重ねた画面から戻る操作。積まれた画面の中でだけ値を持つ
export const MobileBackContext = createContext<(() => void) | null>(null);

export const useMobileBack = () => useContext(MobileBackContext);
