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
  const { mutate: updateReadState } = useUpdateReadState(workspaceId);
  // 既読にした対象。チャンネルや最新のメッセージが変わったらまた既読にする
  const markedKeyRef = useRef<string | null>(null);

  useEffect(() => {
    if (element === null || channelId === null || latestMessageId === null) {
      return undefined;
    }

    const key = `${channelId}:${latestMessageId}:${String(includeDescendants)}`;
    const observer = new IntersectionObserver(
      (entries) => {
        const [entry] = entries;
        if (entry?.isIntersecting && markedKeyRef.current !== key) {
          markedKeyRef.current = key;
          updateReadState({
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
  }, [element, channelId, latestMessageId, includeDescendants, updateReadState]);

  return { latestMessageRef: setElement };
};
