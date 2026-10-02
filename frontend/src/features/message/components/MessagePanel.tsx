import { useCallback, useMemo } from "react";

import { IconHash } from "@tabler/icons-react";
import { useNavigate, useParams, useSearch } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { ChannelChip } from "#/features/channel/components/ChannelChip";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";
import { useChannelThreadMetadata } from "#/features/message/hooks/useChannelThreadMetadata";
import { useChannelTimeline } from "#/features/message/hooks/useChannelTimeline";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useMessagePages } from "#/features/message/hooks/useMessagePages";
import { useMessageViewportDetection } from "#/features/message/hooks/useMessageViewportDetection";
import { startOfDateKey } from "#/features/message/utils/dateJump";
import { buildTimelineRows } from "#/features/message/utils/timelineRows";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";
import { myUserIdAtom } from "#/providers/store/auth";
import { preferencesAtom } from "#/providers/store/preferences";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useMessages } from "../hooks/useMessage";
import { MessageItem } from "./MessageItem";
import { MessageList } from "./MessageList";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessagePanelProps = {
  workspaceId: string;
  channelId: string;
};

export const MessagePanel = ({ workspaceId, channelId }: MessagePanelProps) => {
  const { t } = useTranslation();
  const myId = useAtomValue(myUserIdAtom);
  const { channel, descendants, includesDescendants, isResolved } = useChannelAggregation(
    workspaceId,
    channelId,
  );
  const jumpDate = useSearch({
    from: "/app/$workspaceId/$channelId",
    select: (search) => search.date ?? null,
  });
  const messageParam = useSearch({
    from: "/app/$workspaceId/$channelId",
    select: (search) => search.message ?? null,
  });
  // スレッドを開いているときの ?message= はスレッドの返信を指すため、チャンネルでは扱わない
  const isThreadOpen = useParams({
    select: (params) => params.messageId !== undefined,
    strict: false,
  });
  const { timeZone } = useDateFormat();
  const around = jumpDate === null ? null : startOfDateKey(jumpDate, timeZone);
  const {
    data: messageResponse,
    isLoading: isLoadingMessages,
    isError,
    error,
  } = useMessages(isResolved ? channelId : null, includesDescendants, around);
  const isLoading = !isResolved || isLoadingMessages;
  const { wsClient } = useWsClient();
  const threadMetadataById = useChannelThreadMetadata(
    isResolved ? channelId : null,
    includesDescendants,
  );

  const {
    items: initialMessages,
    hasMore: hasOlderMessages,
    hasNewer: hasNewerMessages,
    load,
    loading,
  } = useMessagePages({
    base: messageResponse,
    channelId,
    includeDescendants: includesDescendants,
    jumpDate,
  });

  // 一覧にない未参加の公開子孫も、届いたメッセージから購読する
  const descendantIds = includesDescendants
    ? [
        ...descendants.map((descendant) => descendant.id),
        ...(initialMessages ?? []).flatMap((item) =>
          item.content.case === "userMessage" && item.content.value.channelId !== channelId
            ? [item.content.value.channelId]
            : [],
        ),
      ]
    : [];

  const { orderedItems } = useChannelTimeline({
    currentChannelId: channelId,
    descendantIds,
    initialMessages,
    wsClient: wsClient ?? null,
  });

  const { hideJoinMessages } = useAtomValue(preferencesAtom);
  const rows = useMemo(
    () => buildTimelineRows(orderedItems, hideJoinMessages, timeZone),
    [orderedItems, hideJoinMessages, timeZone],
  );
  const navigate = useNavigate();

  // 最新メッセージのIDを取得（ユーザーメッセージのみ）
  const latestUserMessageId =
    orderedItems.length > 0
      ? (() => {
          for (let i = orderedItems.length - 1; i >= 0; i--) {
            const item = orderedItems[i];
            if (item?.content.case === "userMessage") {
              return item.content.value.id;
            }
          }
          return null;
        })()
      : null;

  const { latestMessageRef } = useMessageViewportDetection({
    channelId,
    includeDescendants: includesDescendants,
    // 日付へ移動して最新まで読み込んでいないうちは既読にしない
    latestMessageId: hasNewerMessages ? null : latestUserMessageId,
    workspaceId,
  });

  // 指定日以降で最初の投稿。一覧は新しい順に並ぶ
  const jumpTargetId =
    (around &&
      messageResponse?.messages.findLast((item) => toDate(item.createdAt) >= around)?.content.value
        ?.id) ??
    null;

  const handleCopyLink = useCopyMessageLink(workspaceId, channelId);

  const handleOpenThread = useCallback(
    (messageId: string) => {
      void navigate({
        params: { channelId, messageId, workspaceId },
        to: "/app/$workspaceId/$channelId/thread/$messageId",
      });
    },
    [navigate, channelId, workspaceId],
  );

  const renderMessage = (msg: Message, isHighlighted: boolean) => (
    <MessageItem
      message={msg}
      currentUserId={myId}
      onCopyLink={handleCopyLink}
      onCreateThread={handleOpenThread}
      onOpenThread={handleOpenThread}
      threadMetadata={threadMetadataById.get(msg.id)}
      isHighlighted={isHighlighted}
      channelChip={
        channel && msg.channelId !== channel.id ? (
          <ChannelChip
            workspaceId={workspaceId}
            parentName={channel.name}
            channelId={msg.channelId}
          />
        ) : null
      }
    />
  );

  const renderBody = () => {
    if (isLoading) {
      return (
        <div className="flex flex-1 flex-col justify-end gap-4 px-[18px] py-4" aria-busy>
          {[0, 1, 2].map((index) => (
            <div key={index} className="flex gap-2.5">
              <Skeleton className="size-8 shrink-0 rounded-md" />
              <div className="flex flex-1 flex-col gap-1.5">
                <Skeleton className="h-3.5 w-32" />
                <Skeleton className="h-3.5 w-3/4" />
              </div>
            </div>
          ))}
        </div>
      );
    }
    if (isError) {
      return <p className="m-0 px-[18px] py-4 text-body text-danger">{error.message}</p>;
    }
    if (orderedItems.length === 0) {
      return (
        <div className="flex flex-1 flex-col items-center justify-center gap-1.5 p-6 text-center">
          <span className="mb-1 grid size-11 place-items-center rounded-lg bg-sunken text-muted [&_svg]:size-[22px]">
            <IconHash aria-hidden />
          </span>
          <b className="text-body-strong">{t("message.panel.empty")}</b>
        </div>
      );
    }
    return (
      <MessageList
        // チャンネルや日付を切り替えたら、位置と読み込み状態を作り直す
        key={`${channelId}:${String(includesDescendants)}:${jumpDate ?? ""}`}
        rows={rows}
        currentUserId={myId}
        targetMessageId={isThreadOpen ? null : (messageParam ?? jumpTargetId)}
        hasOlder={hasOlderMessages}
        hasNewer={hasNewerMessages}
        loading={loading}
        onLoad={(direction) => {
          void load(direction);
        }}
        onJumpToLatest={() => {
          void navigate({ search: (prev) => ({ ...prev, date: undefined }), to: "." });
        }}
        latestMessageRef={latestMessageRef}
        latestUserMessageId={latestUserMessageId}
        renderMessage={renderMessage}
        header={null}
      />
    );
  };

  return <div className="flex h-full min-h-0 w-full flex-col bg-surface">{renderBody()}</div>;
};
