import { useEffect, useRef, useState } from "react";

const HIGHLIGHT_DURATION_MS = 3_000;

/**
 * TargetMessageId のメッセージへスクロールし、一定時間ハイライトする。 scrollToMessage は対象がまだ一覧にないとき false
 * を返し、一覧が変わって作り直されたときに再び試す。
 */
export const useHighlightedMessage = (
  isReady: boolean,
  targetMessageId: string | null,
  scrollToMessage: (messageId: string) => boolean,
) => {
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

  return isHighlightExpired ? null : targetMessageId;
};
