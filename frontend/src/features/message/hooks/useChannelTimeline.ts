import { useEffect, useMemo, useState } from "react";

import type { TimelineItem } from "#/features/message/types";
import type { NewMessagePayload, SystemMessageCreatedPayload } from "#/types/wsEvents";

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
    setTimeline(initialMessages ?? []);
  }, [initialMessages, currentChannelId]);

  // WS 購読と join/leave 管理
  useEffect(() => {
    if (!wsClient || !currentChannelId) {
      return undefined;
    }
    wsClient.joinChannel(currentChannelId);

    const handleNewMessage = ({ message }: NewMessagePayload) => {
      setTimeline((prev) => {
        if (prev.some((item) => item.type === "user" && item.userMessage?.id === message.id)) {
          return prev;
        }
        return [...prev, { createdAt: message.createdAt, type: "user", userMessage: message }];
      });
    };

    const handleSystem = ({ message }: SystemMessageCreatedPayload) => {
      setTimeline((prev) => {
        if (prev.some((item) => item.type === "system" && item.systemMessage?.id === message.id)) {
          return prev;
        }
        return [...prev, { createdAt: message.createdAt, systemMessage: message, type: "system" }];
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
    if (!currentChannelId) {
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
            (i) => i.type === "system" && i.systemMessage?.id === item.systemMessage?.id,
          )
        );
      }
      return true;
    });
    return unique.toSorted((a: TimelineItem, b: TimelineItem) => {
      const at = new Date(a.createdAt).getTime();
      const bt = new Date(b.createdAt).getTime();
      return at - bt;
    });
  }, [timeline, currentChannelId]);

  return { orderedItems };
};
