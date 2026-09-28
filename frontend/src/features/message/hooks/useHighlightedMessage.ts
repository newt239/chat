import { useCallback, useEffect, useRef, useState } from "react";

import { useSearch } from "@tanstack/react-router";

const HIGHLIGHT_DURATION_MS = 3_000;

/** ?message=<id> で指定されたメッセージへスクロールし、一定時間ハイライトする。 対象が描画された時点で ref が渡るため、ref コールバックでスクロールする。 */
export const useHighlightedMessage = (isReady: boolean) => {
  const targetMessageId = useSearch({
    from: "/app/$workspaceId/$channelId",
    select: (search) => search.message ?? null,
  });

  const [isHighlightExpired, setIsHighlightExpired] = useState(false);
  const scrolledMessageIdRef = useRef<string | null>(null);

  useEffect(() => {
    setIsHighlightExpired(false);
    const timer = setTimeout(() => {
      setIsHighlightExpired(true);
    }, HIGHLIGHT_DURATION_MS);

    return () => {
      clearTimeout(timer);
    };
  }, [targetMessageId]);

  const highlightedMessageRef = useCallback(
    (element: HTMLDivElement | null) => {
      if (element === null || !isReady || targetMessageId === null) {
        return;
      }
      // 再レンダリングのたびにスクロールし直さない
      if (scrolledMessageIdRef.current === targetMessageId) {
        return;
      }
      scrolledMessageIdRef.current = targetMessageId;
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    },
    [isReady, targetMessageId],
  );

  return {
    highlightedId: isHighlightExpired ? null : targetMessageId,
    highlightedMessageRef,
    targetMessageId,
  };
};
