import { useEffect, useRef, useState } from "react";

import { useParams, useSearch } from "@tanstack/react-router";

const HIGHLIGHT_DURATION_MS = 3_000;

/**
 * ?message=<id> で指定されたメッセージ（無ければ jumpTargetId）へスクロールし、一定時間ハイライトする。 scrollToMessage は対象がまだ一覧にないとき
 * false を返し、一覧が変わって作り直されたときに再び試す。
 */
export const useHighlightedMessage = (
  isReady: boolean,
  jumpTargetId: string | null,
  scrollToMessage: (messageId: string) => boolean,
) => {
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

  useEffect(() => {
    if (!isReady || targetMessageId === null) {
      return;
    }
    // 再レンダリングのたびにスクロールし直さない
    if (scrolledMessageIdRef.current === targetMessageId) {
      return;
    }
    if (scrollToMessage(targetMessageId)) {
      scrolledMessageIdRef.current = targetMessageId;
    }
  }, [isReady, targetMessageId, scrollToMessage]);

  return {
    highlightedId: isHighlightExpired ? null : targetMessageId,
    targetMessageId,
  };
};
