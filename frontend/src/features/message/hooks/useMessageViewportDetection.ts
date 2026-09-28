import { useEffect, useRef } from "react";

import { timestampNow } from "@bufbuild/protobuf/wkt";

import { useUpdateReadState } from "#/features/channel/hooks/useUpdateReadState";

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
  const updateReadState = useUpdateReadState(workspaceId);
  const updateReadStateRef = useRef(updateReadState);
  const hasMarkedAsRead = useRef(false);

  useEffect(() => {
    updateReadStateRef.current = updateReadState;
  }, [updateReadState]);

  useEffect(() => {
    hasMarkedAsRead.current = false;

    if (latestMessageRef.current === null || channelId === null || latestMessageId === null) {
      return undefined;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (entry?.isIntersecting && !hasMarkedAsRead.current) {
          hasMarkedAsRead.current = true;
          updateReadStateRef.current.mutate({ channelId, lastReadAt: timestampNow() });
        }
      },
      {
        rootMargin: "0px",
        threshold: 0.1,
      },
    );

    observer.observe(latestMessageRef.current);

    return () => {
      observer.disconnect();
    };
  }, [channelId, latestMessageId]);

  return { latestMessageRef };
};
