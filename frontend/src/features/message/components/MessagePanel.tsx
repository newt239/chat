import { useCallback, useMemo } from "react";

import { IconHash } from "@tabler/icons-react";
import { useNavigate, useParams, useSearch } from "@tanstack/react-router";
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
import { usePreferences } from "#/hooks/usePreferences";
import { toDate } from "#/lib/timestamp";

import { MessageItem } from "./MessageItem";
import { MessageList } from "./MessageList";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessagePanelProps = {
  workspaceId: string;
  channelId: string;
};

export const MessagePanel = ({ workspaceId, channelId }: MessagePanelProps) => {
  const { t } = useTranslation();
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
    items,
    hasOlder,
    hasNewer,
    load,
    loading,
    isLoading: isLoadingMessages,
    isError,
    error,
  } = useMessagePages({
    around,
    channelId: isResolved ? channelId : null,
    includeDescendants: includesDescendants,
  });
  const isLoading = !isResolved || isLoadingMessages;
  const threadMetadataById = useChannelThreadMetadata(
    isResolved ? channelId : null,
    includesDescendants,
  );

  // 一覧にない未参加の公開子孫も、届いたメッセージから購読する
  const descendantIds = includesDescendants
    ? [
        ...descendants.map((descendant) => descendant.id),
        ...(items ?? []).flatMap((item) =>
          item.content.case === "userMessage" && item.content.value.channelId !== channelId
            ? [item.content.value.channelId]
            : [],
        ),
      ]
    : [];

  const { orderedItems } = useChannelTimeline({
    channelId,
    descendantIds,
    includeDescendants: includesDescendants,
    items,
  });

  const { hideJoinMessages } = usePreferences();
  const rows = useMemo(
    () => buildTimelineRows(orderedItems, hideJoinMessages, timeZone),
    [orderedItems, hideJoinMessages, timeZone],
  );
  const navigate = useNavigate();

  const latestUserMessage = orderedItems.findLast((item) => item.content.case === "userMessage");
  const latestUserMessageId = latestUserMessage?.content.value?.id ?? null;

  const { latestMessageRef } = useMessageViewportDetection({
    channelId,
    includeDescendants: includesDescendants,
    // 日付へ移動して最新まで読み込んでいないうちは既読にしない
    latestMessageId: hasNewer ? null : latestUserMessageId,
    workspaceId,
  });

  // 指定日以降で最初の投稿
  const jumpTargetId =
    (around && orderedItems.find((item) => toDate(item.createdAt) >= around)?.content.value?.id) ??
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
      onCopyLink={handleCopyLink}
      onCreateThread={handleOpenThread}
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
      return <p className="m-0 px-[18px] py-4 text-body text-danger">{error?.message}</p>;
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
        targetMessageId={isThreadOpen ? null : (messageParam ?? jumpTargetId)}
        hasOlder={hasOlder}
        hasNewer={hasNewer}
        loading={loading}
        onLoad={load}
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
