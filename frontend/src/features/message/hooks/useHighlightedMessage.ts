import { useCallback, useEffect, useRef, useState } from "react";

import { useParams, useSearch } from "@tanstack/react-router";

const HIGHLIGHT_DURATION_MS = 3_000;

/**
 * ?message=<id> で指定されたメッセージ（無ければ jumpTargetId）へスクロールし、一定時間ハイライトする。 対象が描画された時点で ref が渡るため、ref
 * コールバックでスクロールする。
 */
export const useHighlightedMessage = (isReady: boolean, jumpTargetId: string | null) => {
  // スレッドを開いているときの ?message= はスレッドの返信を指すため、チャンネルでは扱わない
  const isThreadOpen = useParams({
    select: (params) => params.messageId !== undefined,
    strict: false,
  });
  const message = useSearch({
    from: "/app/$workspaceId/$channelId",
    select: (search) => search.message ?? null,
  });
  const targetMessageId = isThreadOpen ? null : (message ?? jumpTargetId);

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
