import { useEffect } from "react";

import { create } from "@bufbuild/protobuf";
import { useQueryClient } from "@tanstack/react-query";
import { useAtomValue } from "jotai";

import {
  subscribeMessagePatches,
  updateMessagePages,
  updateTimelineMessage,
  updateUserMessages,
} from "#/features/message/utils/updateTimelineMessage";
import { pinListKey } from "#/features/pin/hooks/useTogglePin";
import { ThreadMetadataSchema, TimelineItemSchema } from "#/gen/chat/v1/message_pb";
import { myUserIdAtom } from "#/providers/store/auth";
import { useWsClient } from "#/providers/ws/useWsClient";

import { messagePagesKey } from "./useMessagePages";

import type { TimelineItem } from "#/gen/chat/v1/message_pb";

type UseChannelTimelineArgs = {
  channelId: string;
  includeDescendants: boolean;
  // 集約表示中の子孫チャンネル。購読してタイムラインに新着を積む
  descendantIds: readonly string[];
  // useMessagePages の項目（新しい順）
  items: TimelineItem[] | undefined;
};

/** WebSocket の差分を読み込み済みのページに当て、表示用に古い順へ並べ直す */
export const useChannelTimeline = ({
  channelId,
  includeDescendants,
  descendantIds,
  items,
}: UseChannelTimelineArgs) => {
  const queryClient = useQueryClient();
  const wsClient = useWsClient();
  const myId = useAtomValue(myUserIdAtom);
  const descendantKey = [...new Set(descendantIds)].toSorted().join(",");

  useEffect(() => {
    if (!wsClient) {
      return undefined;
    }
    const channelIds = [channelId, ...descendantKey.split(",").filter(Boolean)];
    for (const id of channelIds) {
      wsClient.joinChannel(id);
    }

    const updatePages = (
      eventChannelId: string,
      update: (items: TimelineItem[], isLatestPage: boolean) => TimelineItem[],
    ) => {
      if (!channelIds.includes(eventChannelId)) {
        return;
      }
      updateMessagePages(queryClient, messagePagesKey(channelId, includeDescendants), update);
    };

    const prependToLatest = (eventChannelId: string, item: TimelineItem) => {
      const id = item.content.value?.id;
      updatePages(eventChannelId, (current, isLatestPage) =>
        isLatestPage && !current.some((existing) => existing.content.value?.id === id)
          ? [item, ...current]
          : current,
      );
    };

    // ピンの一覧も取り直す。自分の操作は useTogglePin で取り直すため二重に取らない
    const invalidatePins = (eventChannelId: string, pinnedBy: string) => {
      if (pinnedBy !== myId) {
        void queryClient.invalidateQueries({ queryKey: pinListKey(eventChannelId) });
      }
    };

    const unsubscribes = [
      wsClient.on("newMessage", ({ channelId: eventChannelId, message }) => {
        if (message === undefined) {
          return;
        }
        // スレッドの返信はタイムラインに積まず、親の返信数などを進める。返信者と親の投稿者は自動でフォローされる
        const { parentId } = message;
        if (parentId !== undefined) {
          updatePages(eventChannelId, (current) =>
            updateTimelineMessage(current, parentId, (parent) => ({
              ...parent,
              threadMetadata: create(ThreadMetadataSchema, {
                isFollowing:
                  (parent.threadMetadata?.isFollowing ?? false) ||
                  message.userId === myId ||
                  parent.userId === myId,
                lastReplyAt: message.createdAt,
                lastReplyUser: message.user,
                replyCount: (parent.threadMetadata?.replyCount ?? 0) + 1,
              }),
            })),
          );
          return;
        }
        prependToLatest(
          eventChannelId,
          create(TimelineItemSchema, {
            content: { case: "userMessage", value: message },
            createdAt: message.createdAt,
          }),
        );
      }),

      wsClient.on("systemMessageCreated", ({ channelId: eventChannelId, message }) => {
        if (message !== undefined) {
          prependToLatest(
            eventChannelId,
            create(TimelineItemSchema, {
              content: { case: "systemMessage", value: message },
              createdAt: message.createdAt,
            }),
          );
        }
      }),

      wsClient.on("pinCreated", ({ channelId: eventChannelId, pinnedBy }) => {
        invalidatePins(eventChannelId, pinnedBy);
      }),
      wsClient.on("pinDeleted", ({ channelId: eventChannelId, pinnedBy }) => {
        invalidatePins(eventChannelId, pinnedBy);
      }),
      subscribeMessagePatches(wsClient, (eventChannelId, messageIds, update) => {
        updatePages(eventChannelId, (current) =>
          updateUserMessages(current, (message) => messageIds.has(message.id), update),
        );
      }),
    ];

    return () => {
      for (const unsubscribe of unsubscribes) {
        unsubscribe();
      }
      for (const id of channelIds) {
        wsClient.leaveChannel(id);
      }
    };
  }, [wsClient, queryClient, channelId, includeDescendants, descendantKey, myId]);

  // ページの境目で重なった項目を除き、古い順にする
  const seen = new Set<string>();
  const orderedItems = (items ?? [])
    .filter((item) => {
      const id = item.content.value?.id;
      if (id === undefined || seen.has(id)) {
        return id === undefined;
      }
      seen.add(id);
      return true;
    })
    .toReversed();

  return { orderedItems };
};
