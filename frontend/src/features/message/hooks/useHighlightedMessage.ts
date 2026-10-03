import { useEffect, useRef, useState } from "react";

const HIGHLIGHT_DURATION_MS = 3_000;

// 対象へスクロールして一定時間ハイライトする。scrollToMessage が false なら作り直されたときに再び試す
export const useHighlightedMessage = (
  targetMessageId: string | null,
  scrollToMessage: (messageId: string) => boolean,
) => {
  const [expiredMessageId, setExpiredMessageId] = useState<string | null>(null);
  const scrolledMessageIdRef = useRef<string | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setExpiredMessageId(targetMessageId);
    }, HIGHLIGHT_DURATION_MS);

    return () => {
      clearTimeout(timer);
    };
  }, [targetMessageId]);

  useEffect(() => {
    if (targetMessageId === null) {
      return;
    }
    // 再レンダリングのたびにスクロールし直さない
    if (scrolledMessageIdRef.current === targetMessageId) {
      return;
    }
    if (scrollToMessage(targetMessageId)) {
      scrolledMessageIdRef.current = targetMessageId;
    }
  }, [targetMessageId, scrollToMessage]);

  return expiredMessageId === targetMessageId ? null : targetMessageId;
};
