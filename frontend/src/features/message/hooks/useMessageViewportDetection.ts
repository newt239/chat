import { useEffect, useRef } from "react";

import { useUpdateReadState } from "./useMessage";

type UseMessageViewportDetectionArgs = {
  channelId: string | null;
  workspaceId: string | null;
  latestMessageId: string | null;
};

export const useMessageViewportDetection = ({
  channelId,
  workspaceId,
  latestMessageId,
}: UseMessageViewportDetectionArgs) => {
  const latestMessageRef = useRef<HTMLDivElement | null>(null);
  const updateReadState = useUpdateReadState(channelId, workspaceId);
  const updateReadStateRef = useRef(updateReadState);
  const hasMarkedAsRead = useRef(false);

  useEffect(() => {
    updateReadStateRef.current = updateReadState;
  }, [updateReadState]);

  useEffect(() => {
    hasMarkedAsRead.current = false;

    if (latestMessageRef.current === null || channelId === null || latestMessageId === null) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry?.isIntersecting && !hasMarkedAsRead.current) {
          hasMarkedAsRead.current = true;
          updateReadStateRef.current.mutate(undefined);
        }
      },
      {
        threshold: 0.1,
        rootMargin: "0px",
      }
    );

    observer.observe(latestMessageRef.current);

    return () => {
      observer.disconnect();
    };
  }, [channelId, latestMessageId]);

  return { latestMessageRef };
};
