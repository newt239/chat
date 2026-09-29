import { useCallback, useSyncExternalStore } from "react";

export const useMediaQuery = (query: string) => {
  const subscribe = useCallback(
    (onChange: () => void) => {
      const media = globalThis.matchMedia(query);
      media.addEventListener("change", onChange);
      return () => {
        media.removeEventListener("change", onChange);
      };
    },
    [query],
  );
  return useSyncExternalStore(subscribe, () => globalThis.matchMedia(query).matches);
};

// Tailwind の md ブレークポイント未満をモバイルとして扱う
export const useIsMobile = () => useMediaQuery("(max-width: 767px)");
