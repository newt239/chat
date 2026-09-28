import { useEffect, useMemo, useRef, useCallback } from "react";

import { IconHash } from "@tabler/icons-react";
import { useAtom, useSetAtom, useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { Skeleton } from "#/components/ui/Skeleton";
import { ChannelChip } from "#/features/channel/components/ChannelChip";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";
import { useAutoScrollToBottom } from "#/features/message/hooks/useAutoScrollToBottom";
import { useChannelThreadMetadata } from "#/features/message/hooks/useChannelThreadMetadata";
import { useChannelTimeline } from "#/features/message/hooks/useChannelTimeline";
import { useCopyMessageLink } from "#/features/message/hooks/useCopyMessageLink";
import { useHighlightedMessage } from "#/features/message/hooks/useHighlightedMessage";
import { useMessageViewportDetection } from "#/features/message/hooks/useMessageViewportDetection";
import { useOlderMessages } from "#/features/message/hooks/useOlderMessages";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";
import { currentChannelIdAtom, currentWorkspaceIdAtom } from "#/providers/store/workspace";
import { useWsClient } from "#/providers/ws/useWsClient";

import { useMessages } from "../hooks/useMessage";
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
  const {
    data: messageResponse,
    isLoading: isLoadingMessages,
    isError,
    error,
  } = useMessages(isResolved ? currentChannelId : null, includesDescendants);
  const isLoading = !isResolved || isLoadingMessages;
  const { wsClient } = useWsClient();
  const threadMetadataById = useChannelThreadMetadata(
    isResolved ? currentChannelId : null,
    includesDescendants,
  );

  const {
    olderItems,
    hasMore: hasOlderMessages,
    isLoading: isLoadingOlder,
    loadOlder,
  } = useOlderMessages(currentChannelId, includesDescendants, messageResponse?.hasMore ?? false);

  const initialMessages = useMemo(
    () => (messageResponse ? [...olderItems, ...messageResponse.messages] : undefined),
    [olderItems, messageResponse],
  );

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

  const handleLoadOlder = () => {
    const oldest = orderedItems.at(0);
    if (oldest) {
      void loadOlder(oldest.createdAt);
    }
  };

  const messagesEndRef = useRef<HTMLDivElement>(null);
  const setRightSidebarView = useSetAtom(setRightSidePanelViewAtom);

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
    latestMessageId: latestUserMessageId,
    workspaceId: currentWorkspaceId,
  });

  const scrollToBottom = useAutoScrollToBottom(messagesEndRef);
  const { highlightedId, highlightedMessageRef, targetMessageId } =
    useHighlightedMessage(!isLoading);

  useEffect(() => {
    // 特定メッセージへのリンクで開いたときは最下部へ飛ばさない
    if (targetMessageId === null) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading, currentChannelId, scrollToBottom, targetMessageId]);

  useEffect(() => {
    setRightSidebarView({ type: "hidden" });
  }, [currentChannelId, setRightSidebarView]);

  const handleCopyLink = useCopyMessageLink(currentWorkspaceId, currentChannelId);

  const handleOpenThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ threadId: messageId, type: "thread" });
    },
    [setRightSidebarView],
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
            <Button variant="ghost" size="sm" isPending={isLoadingOlder} onPress={handleLoadOlder}>
              {t("message.panel.loadOlder")}
            </Button>
          </div>
        )}
        {/* ホバー時のツールバーがメッセージの上にはみ出すため、先頭に余白を取る */}
        <div className="flex flex-1 flex-col justify-end pt-8 pb-2">
          {orderedItems.map((item) => {
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
              return (
                <SystemMessageItem
                  key={`s-${item.content.value.id}`}
                  message={item.content.value}
                />
              );
            }
            return null;
          })}
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
