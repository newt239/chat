import { useRef } from "react";

import { timestampNow } from "@bufbuild/protobuf/wkt";

import { useUpdateReadState } from "#/features/channel/hooks/useUpdateReadState";

type UseMessageViewportDetectionArgs = {
  channelId: string;
  workspaceId: string;
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
  const { mutate: updateReadState } = useUpdateReadState(workspaceId);
  // 既読にした対象。チャンネルや最新のメッセージが変わったらまた既読にする
  const markedKeyRef = useRef<string | null>(null);

  const latestMessageRef = (element: HTMLElement | null) => {
    if (element === null || latestMessageId === null) {
      return undefined;
    }
    const key = `${channelId}:${latestMessageId}:${String(includeDescendants)}`;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry?.isIntersecting && markedKeyRef.current !== key) {
          markedKeyRef.current = key;
          updateReadState({ channelId, includeDescendants, lastReadAt: timestampNow() });
        }
      },
      { threshold: 0.1 },
    );
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  };

  return { latestMessageRef };
};
