import { useCallback } from "react";

export const useAutoScrollToBottom = (ref: React.RefObject<HTMLDivElement | null>) =>
  useCallback(() => {
    ref.current?.scrollIntoView({ behavior: "smooth" });
  }, [ref]);
