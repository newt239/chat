import { useEffect, useRef, useState } from "react";

import { timestampNow } from "@bufbuild/protobuf/wkt";

import { useUpdateReadState } from "#/features/channel/hooks/useUpdateReadState";

type UseMessageViewportDetectionArgs = {
  channelId: string | null;
  workspaceId: string | null;
  latestMessageId: string | null;
  // 集約表示中は子孫チャンネルもまとめて既読にする
  includeDescendants: boolean;
};

/** 最新のメッセージが画面に入ったら既読にする。仮想リストでは行が出入りするため、要素は ref コールバックで受け取る */
export const useMessageViewportDetection = ({
  channelId,
  workspaceId,
  latestMessageId,
  includeDescendants,
}: UseMessageViewportDetectionArgs) => {
  const [element, setElement] = useState<HTMLElement | null>(null);
  const updateReadState = useUpdateReadState(workspaceId);
  const updateReadStateRef = useRef(updateReadState);
  const hasMarkedAsRead = useRef(false);

  useEffect(() => {
    updateReadStateRef.current = updateReadState;
  }, [updateReadState]);

  useEffect(() => {
    hasMarkedAsRead.current = false;
  }, [channelId, latestMessageId, includeDescendants]);

  useEffect(() => {
    if (element === null || channelId === null || latestMessageId === null) {
      return undefined;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (entry?.isIntersecting && !hasMarkedAsRead.current) {
          hasMarkedAsRead.current = true;
          updateReadStateRef.current.mutate({
            channelId,
            includeDescendants,
            lastReadAt: timestampNow(),
          });
        }
      },
      {
        rootMargin: "0px",
        threshold: 0.1,
      },
    );

    observer.observe(element);

    return () => {
      observer.disconnect();
    };
  }, [element, channelId, latestMessageId, includeDescendants]);

  return { latestMessageRef: setElement };
};
