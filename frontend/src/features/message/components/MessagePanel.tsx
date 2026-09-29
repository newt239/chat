import { useEffect, useRef, useCallback } from "react";

import { IconHash } from "@tabler/icons-react";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useAtom, useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { ChannelChip } from "#/features/channel/components/ChannelChip";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";
import { useAutoScrollToBottom } from "#/features/message/hooks/useAutoScrollToBottom";
import { useChannelThreadMetadata } from "#/features/message/hooks/useChannelThreadMetadata";
import { useChannelTimeline } from "#/features/message/hooks/useChannelTimeline";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";
import { useMessagePages } from "#/features/message/hooks/useMessagePages";
import { useMessageViewportDetection } from "#/features/message/hooks/useMessageViewportDetection";
import { groupByDate, startOfDateKey } from "#/features/message/utils/dateJump";
import { toDate } from "#/lib/timestamp";
import { userAtom } from "#/providers/store/auth";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useMessages } from "../hooks/useMessage";
import { DateDivider } from "./DateDivider";
import { MessageItem } from "./MessageItem";
import { SystemMessageItem } from "./SystemMessageItem";
import { TypingIndicator } from "./TypingIndicator";

export const MessagePanel = () => {
  const { t } = useTranslation();
  const [currentWorkspaceId] = useAtom(currentWorkspaceIdAtom);
  const [currentChannelId] = useAtom(currentChannelIdAtom);
  const currentUser = useAtomValue(userAtom);
  const { channel, descendants, includesDescendants, isResolved } = useChannelAggregation(
    currentWorkspaceId,
    currentChannelId,
  );
  const jumpDate = useSearch({
    from: "/app/$workspaceId/$channelId",
    select: (search) => search.date ?? null,
  });
  const around = jumpDate === null ? null : startOfDateKey(jumpDate);
  const {
    data: messageResponse,
    isLoading: isLoadingMessages,
    isError,
    error,
  } = useMessages(isResolved ? currentChannelId : null, includesDescendants, around);
  const isLoading = !isResolved || isLoadingMessages;
  const { wsClient } = useWsClient();
  const threadMetadataById = useChannelThreadMetadata(
    isResolved ? currentChannelId : null,
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
    channelId: currentChannelId,
    includeDescendants: includesDescendants,
    jumpDate,
  });

  // 一覧にない未参加の公開子孫も、届いたメッセージから購読する
  const descendantIds = includesDescendants
    ? [
        ...descendants.map((descendant) => descendant.id),
        ...(initialMessages ?? []).flatMap((item) =>
          item.content.case === "userMessage" && item.content.value.channelId !== currentChannelId
            ? [item.content.value.channelId]
            : [],
        ),
      ]
    : [];

  const { orderedItems, typingUserIds } = useChannelTimeline({
    currentChannelId,
    descendantIds,
    initialMessages,
    wsClient: wsClient ?? null,
  });

  const messagesEndRef = useRef<HTMLDivElement>(null);
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
    channelId: currentChannelId,
    includeDescendants: includesDescendants,
    // 日付へ移動して最新まで読み込んでいないうちは既読にしない
    latestMessageId: hasNewerMessages ? null : latestUserMessageId,
    workspaceId: currentWorkspaceId,
  });

  const scrollToBottom = useAutoScrollToBottom(messagesEndRef);
  // 指定日以降で最初の投稿。一覧は新しい順に並ぶ
  const jumpTargetId =
    (around &&
      messageResponse?.messages.findLast((item) => toDate(item.createdAt) >= around)?.content.value
        ?.id) ??
    null;
  const { highlightedId, highlightedMessageRef, targetMessageId } = useHighlightedMessage(
    !isLoading,
    jumpTargetId,
  );

  // 取得したらタイムラインに反映されるのを待って最下部へスクロールする
  const shouldScrollRef = useRef(false);
  useEffect(() => {
    shouldScrollRef.current = true;
  }, [messageResponse, currentChannelId, targetMessageId]);
  const newestId = messageResponse?.messages.at(0)?.content.value?.id;
  useEffect(() => {
    // 特定メッセージへのリンクや日付で開いたときは最下部へ飛ばさない
    if (!shouldScrollRef.current || targetMessageId !== null) {
      return;
    }
    if (
      newestId !== undefined &&
      !orderedItems.some((item) => item.content.value?.id === newestId)
    ) {
      return;
    }
    shouldScrollRef.current = false;
    scrollToBottom();
  }, [orderedItems, newestId, scrollToBottom, targetMessageId]);

  const handleCopyLink = useCopyMessageLink(currentWorkspaceId, currentChannelId);

  const handleOpenThread = useCallback(
    (messageId: string) => {
      if (currentWorkspaceId !== null && currentChannelId !== null) {
        void navigate({
          params: { channelId: currentChannelId, messageId, workspaceId: currentWorkspaceId },
          to: "/app/$workspaceId/$channelId/thread/$messageId",
        });
      }
    },
    [navigate, currentChannelId, currentWorkspaceId],
  );

  if (currentWorkspaceId === null || currentChannelId === null) {
    return (
      <div className="grid h-full place-items-center p-6 font-sans text-body text-muted">
        {t(
          currentWorkspaceId === null
            ? "message.panel.selectWorkspace"
            : "message.panel.selectChannel",
        )}
      </div>
    );
  }

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
          <span className="text-caption text-muted">{t("message.panel.emptyHint")}</span>
        </div>
      );
    }
    return (
      <>
        {hasOlderMessages && (
          <div className="flex justify-center py-2">
            <Button
              variant="ghost"
              size="sm"
              isPending={loading === "older"}
              onPress={() => {
                void load("older");
              }}
            >
              {t("message.panel.loadOlder")}
            </Button>
          </div>
        )}
        {/* ホバー時のツールバーがメッセージの上にはみ出すため、先頭に余白を取る */}
        <div className="flex flex-1 flex-col justify-end pt-8 pb-2">
          {groupByDate(orderedItems).map(({ dateKey, items }) => (
            <section key={dateKey}>
              <DateDivider dateKey={dateKey} />
              {items.map((item) => {
                if (item.content.case === "userMessage") {
                  const msg = item.content.value;
                  return (
                    <div
                      key={`u-${msg.id}`}
                      ref={
                        msg.id === targetMessageId
                          ? highlightedMessageRef
                          : msg.id === latestUserMessageId
                            ? latestMessageRef
                            : undefined
                      }
                    >
                      <MessageItem
                        message={msg}
                        currentUserId={currentUser?.id ?? null}
                        onCopyLink={handleCopyLink}
                        onCreateThread={handleOpenThread}
                        onOpenThread={handleOpenThread}
                        threadMetadata={threadMetadataById.get(msg.id)}
                        isHighlighted={msg.id === highlightedId}
                        channelChip={
                          channel && msg.channelId !== channel.id ? (
                            <ChannelChip
                              workspaceId={currentWorkspaceId}
                              parentName={channel.name}
                              channelId={msg.channelId}
                            />
                          ) : null
                        }
                      />
                    </div>
                  );
                }
                if (item.content.case === "systemMessage") {
                  const { id } = item.content.value;
                  return (
                    <div
                      key={`s-${id}`}
                      ref={id === targetMessageId ? highlightedMessageRef : undefined}
                    >
                      <SystemMessageItem message={item.content.value} />
                    </div>
                  );
                }
                return null;
              })}
            </section>
          ))}
          {hasNewerMessages && (
            <div className="flex justify-center gap-2 py-2">
              <Button
                variant="ghost"
                size="sm"
                isPending={loading === "newer"}
                onPress={() => {
                  void load("newer");
                }}
              >
                {t("message.panel.loadNewer")}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onPress={() => {
                  void navigate({ search: (prev) => ({ ...prev, date: undefined }), to: "." });
                }}
              >
                {t("message.panel.jumpToLatest")}
              </Button>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>
      </>
    );
  };

  return (
    <div className="flex h-full min-h-0 w-full flex-col bg-surface">
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">{renderBody()}</div>
      <TypingIndicator userIds={typingUserIds} />
    </div>
  );
};
