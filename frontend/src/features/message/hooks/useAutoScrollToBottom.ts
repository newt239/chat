import { useEffect } from "react";

export const useAutoScrollToBottom = (
  ref: React.RefObject<HTMLDivElement | null>,
  triggers: unknown[]
) => {
  const scrollToBottom = () => {
    ref.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
     
  }, triggers);

  return { scrollToBottom };
};


