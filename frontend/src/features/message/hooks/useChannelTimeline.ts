import { useEffect, useMemo, useState } from "react";


import type { TimelineItem } from "@/features/message/schemas";
import type { NewMessagePayload, SystemMessageCreatedPayload } from "@/types/wsEvents";

import { messageWithThreadSchema, timelineItemSchema } from "@/features/message/schemas";
type WsClientMinimal = {
  joinChannel: (channelId: string) => void;
  leaveChannel: (channelId: string) => void;
  onNewMessage: (cb: (payload: NewMessagePayload) => void) => void;
  offNewMessage: (cb: (payload: NewMessagePayload) => void) => void;
  onSystemMessageCreated: (cb: (payload: SystemMessageCreatedPayload) => void) => void;
  offSystemMessageCreated: (cb: (payload: SystemMessageCreatedPayload) => void) => void;
};

type UseChannelTimelineArgs = {
  currentChannelId: string | null;
  wsClient: WsClientMinimal | null;
  initialMessages: TimelineItem[] | undefined;
};

export const useChannelTimeline = ({
  currentChannelId,
  wsClient,
  initialMessages,
}: UseChannelTimelineArgs) => {
  const [timeline, setTimeline] = useState<TimelineItem[]>([]);

  // 初期ロード・チャンネル変更時に初期化
  useEffect(() => {
    setTimeline((initialMessages ?? []) as TimelineItem[]);
  }, [initialMessages, currentChannelId]);

  // WS 購読と join/leave 管理
  useEffect(() => {
    if (!wsClient || !currentChannelId) return;
    wsClient.joinChannel(currentChannelId);

    const handleNewMessage = (payload: NewMessagePayload) => {
      const result = messageWithThreadSchema.safeParse(payload.message);
      if (!result.success) return;
      setTimeline((prev: TimelineItem[]): TimelineItem[] => {
        const exists = prev.some((m) => m.type === "user" && m.userMessage?.id === result.data.id);
        if (exists) return prev;
        return [
          ...prev,
          { type: "user", userMessage: result.data, createdAt: result.data.createdAt },
        ];
      });
    };

    const handleSystem = (payload: SystemMessageCreatedPayload) => {
      const parsed = timelineItemSchema.shape.systemMessage.unwrap().safeParse(payload.message);
      if (!parsed.success) return;
      const sys = parsed.data;
      setTimeline((prev) => {
        const exists = prev.some((i) => i.type === "system" && i.systemMessage?.id === sys.id);
        if (exists) return prev;
        return [...prev, { type: "system", systemMessage: sys, createdAt: sys.createdAt }];
      });
    };

    wsClient.onNewMessage(handleNewMessage);
    wsClient.onSystemMessageCreated(handleSystem);
    return () => {
      wsClient.offNewMessage(handleNewMessage);
      wsClient.offSystemMessageCreated(handleSystem);
      wsClient.leaveChannel(currentChannelId);
    };
  }, [wsClient, currentChannelId]);

  const orderedItems = useMemo(() => {
    if (!timeline || !currentChannelId) {
      return [] as TimelineItem[];
    }
    const unique = timeline.filter((item: TimelineItem, index: number, self: TimelineItem[]) => {
      if (item.type === "user" && item.userMessage) {
        return (
          index ===
          self.findIndex((i) => i.type === "user" && i.userMessage?.id === item.userMessage?.id)
        );
      }
      if (item.type === "system" && item.systemMessage) {
        return (
          index ===
          self.findIndex(
            (i) => i.type === "system" && i.systemMessage?.id === item.systemMessage?.id
          )
        );
      }
      return true;
    });
    return unique.sort((a: TimelineItem, b: TimelineItem) => {
      const at = new Date(a.createdAt).getTime();
      const bt = new Date(b.createdAt).getTime();
      return at - bt;
    });
  }, [timeline, currentChannelId]);

  return { orderedItems };
};


