import { createContext, useContext } from "react";

import type { ColorMode } from "@chat/design-tokens/theme";

// system を解決した後のモード。トークンを JS で計算する部品（Avatar など）が参照する
export const ColorModeContext = createContext<ColorMode>("light");

export const useColorMode = () => useContext(ColorModeContext);
