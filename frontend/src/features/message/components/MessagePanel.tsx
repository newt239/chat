import { useMemo, useState, useEffect, useRef, useCallback } from "react";
import type { FormEvent } from "react";

import { Card, Loader, Stack, Text, Textarea, ActionIcon } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { IconInfoCircle } from "@tabler/icons-react";
import { useAtom, useSetAtom, useAtomValue } from "jotai";

import { useMessages, useSendMessage, useUpdateMessage, useDeleteMessage } from "../hooks/useMessage";
import { useMessageInputMode } from "../hooks/useMessageInputMode";

import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessageItem } from "./MessageItem";
import { MessagePreview } from "./MessagePreview";
import { ThreadSidePanel } from "./ThreadSidePanel";

import { AttachmentList } from "@/features/attachment/components/AttachmentList";
import { FileInput } from "@/features/attachment/components/FileInput";
import { useFileUpload } from "@/features/attachment/hooks/useFileUpload";
import { useChannels } from "@/features/channel/hooks/useChannel";
import { LinkPreviewCard } from "@/features/link/components/LinkPreviewCard";
import { useLinkPreview } from "@/features/link/hooks/useLinkPreview";
import { userAtom } from "@/lib/store/auth";
import { rightSidebarViewAtom, setRightSidebarViewAtom, toggleRightSidebarViewAtom } from "@/lib/store/ui";

type MessagePanelProps = {
  workspaceId: string | null;
  channelId: string | null;
}

export const MessagePanel = ({ workspaceId, channelId }: MessagePanelProps) => {
  const currentUser = useAtomValue(userAtom);
  const { data: channels } = useChannels(workspaceId);
  const { data: messageResponse, isLoading, isError, error } = useMessages(channelId);
  const sendMessage = useSendMessage(channelId);
  const updateMessage = useUpdateMessage(channelId);
  const deleteMessage = useDeleteMessage(channelId);
  const [body, setBody] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const toggleRightSidebarView = useSetAtom(toggleRightSidebarViewAtom);
  const setRightSidebarView = useSetAtom(setRightSidebarViewAtom);
  const [rightSidebarView] = useAtom(rightSidebarViewAtom);
  const { mode, toggleMode, isEditMode } = useMessageInputMode();
  const linkPreview = useLinkPreview();
  const { previews, addPreview, removePreview, clearPreviews } = linkPreview;
  const fileUpload = useFileUpload();
  const {
    pendingAttachments,
    uploadFile,
    removeAttachment,
    clearAttachments,
    getCompletedAttachmentIds,
    isUploading,
  } = fileUpload;
  // URL検知とリンクプレビューの処理
  const handleBodyChange = useCallback(
    (value: string) => {
      setBody(value);

      // URLを検出してプレビューを追加
      const urlRegex = /https?:\/\/[^\s<>"{}|\\^`\[\]]+/g;
      const urls: string[] = value.match(urlRegex) || [];

      // 新しいURLを検出した場合、プレビューを追加
      urls.forEach((url: string) => {
        if (!previews.some((preview) => preview.url === url)) {
          addPreview(url);
        }
      });

      // 削除されたURLのプレビューを削除
      previews.forEach((preview) => {
        const previewUrl: string = preview.url;
        if (!urls.includes(previewUrl)) {
          removePreview(previewUrl);
        }
      });
    },
    [previews, addPreview, removePreview]
  );

  const dateTimeFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat("ja-JP", {
        dateStyle: "short",
        timeStyle: "short",
      }),
    []
  );
  const orderedMessages = useMemo(() => {
    if (!messageResponse) {
      return [];
    }
    return [...messageResponse.messages].sort((first, second) => {
      const firstTime = new Date(first.createdAt).getTime();
      const secondTime = new Date(second.createdAt).getTime();
      return firstTime - secondTime;
    });
  }, [messageResponse]);

  const activeChannel = useMemo(() => {
    if (!channels || channelId === null) {
      return null;
    }
    return channels.find((channel) => channel.id === channelId) ?? null;
  }, [channels, channelId]);

  // メッセージが読み込まれた時と新しいメッセージが送信された時に最新メッセージにスクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (messageResponse && !isLoading) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading]);

  useEffect(() => {
    if (sendMessage.isSuccess) {
      scrollToBottom();
    }
  }, [sendMessage.isSuccess]);

  // アクションハンドラー
  const handleCopyLink = useCallback(
    (messageId: string) => {
      const url = `${window.location.origin}/app/${workspaceId}/${channelId}?message=${messageId}`;
      navigator.clipboard.writeText(url);
      notifications.show({
        title: "コピーしました",
        message: "メッセージリンクをクリップボードにコピーしました",
      });
    },
    [workspaceId, channelId]
  );

  const handleCreateThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ type: "thread", threadId: messageId });
    },
    [setRightSidebarView]
  );

  const handleOpenThread = useCallback(
    (messageId: string) => {
      setRightSidebarView({ type: "thread", threadId: messageId });
    },
    [setRightSidebarView]
  );

  const handleCloseThread = useCallback(() => {
    setRightSidebarView({ type: "hidden" });
  }, [setRightSidebarView]);

  const handleBookmark = useCallback((messageId: string) => {
    console.log("Bookmark message:", messageId);
    // TODO: ブックマーク機能を実装
  }, []);

  const handleEdit = useCallback(
    (messageId: string, currentBody: string) => {
      updateMessage.mutate(
        { messageId, body: currentBody },
        {
          onSuccess: () => {
            notifications.show({
              title: "更新しました",
              message: "メッセージを更新しました",
            });
          },
          onError: (err) => {
            notifications.show({
              title: "エラー",
              message: err.message ?? "メッセージの更新に失敗しました",
              color: "red",
            });
          },
        }
      );
    },
    [updateMessage]
  );

  const handleDelete = useCallback(
    (messageId: string) => {
      deleteMessage.mutate(
        { messageId },
        {
          onSuccess: () => {
            notifications.show({
              title: "削除しました",
              message: "メッセージを削除しました",
            });
          },
          onError: (err) => {
            notifications.show({
              title: "エラー",
              message: err.message ?? "メッセージの削除に失敗しました",
              color: "red",
            });
          },
        }
      );
    },
    [deleteMessage]
  );

  const handleFileSelect = useCallback(
    async (files: File[]) => {
      if (!channelId) return;

      for (const file of files) {
        await uploadFile(file, { channelId });
      }
    },
    [channelId, uploadFile]
  );

  const isThreadOpen = rightSidebarView.type === "thread";
  const openThreadId = isThreadOpen ? rightSidebarView.threadId : null;

  if (workspaceId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">ワークスペースを選択してください</Text>
      </Card>
    );
  }

  if (channelId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">チャンネルを選択するとメッセージが表示されます</Text>
      </Card>
    );
  }

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault();
    if (body.trim().length === 0 && pendingAttachments.length === 0) {
      return;
    }
    if (isUploading) {
      notifications.show({
        title: "アップロード中",
        message: "ファイルのアップロードが完了するまでお待ちください",
        color: "yellow",
      });
      return;
    }

    const attachmentIds = getCompletedAttachmentIds();
    sendMessage.mutate(
      { body: body.trim(), attachmentIds },
      {
        onSuccess: () => {
          setBody("");
          clearPreviews();
          clearAttachments();
        },
      }
    );
  };

  return (
    <div className="flex h-full min-h-0 flex-col w-full">
      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <div className="flex items-start justify-between">
          <Stack gap="xs" className="flex-1">
            <Text fw={600} size="lg">
              {activeChannel ? `#${activeChannel.name}` : "チャンネル"}
            </Text>
            {activeChannel?.description && (
              <Text size="sm" c="dimmed">
                {activeChannel.description}
              </Text>
            )}
          </Stack>
          <ActionIcon
            variant="subtle"
            color="gray"
            size="lg"
            onClick={() => toggleRightSidebarView({ type: "channel-info", channelId })}
            aria-label="サイドバーの表示切り替え"
          >
            <IconInfoCircle size={20} />
          </ActionIcon>
        </div>
      </Card>

      <div className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error?.message ?? "メッセージの取得に失敗しました"}
          </Text>
        ) : messageResponse && messageResponse.messages.length > 0 ? (
          <div className="flex h-full flex-col">
            {messageResponse.hasMore && (
              <Text size="xs" c="dimmed" className="px-4 py-2 text-center">
                さらに過去のメッセージがあります
              </Text>
            )}
            <div className="flex flex-1 flex-col justify-end">
              <div className="space-y-1 px-4 pb-6">
                {orderedMessages.map((message) => (
                  <MessageItem
                    key={message.id}
                    message={message}
                    currentUserId={currentUser?.id ?? null}
                    dateTimeFormatter={dateTimeFormatter}
                    onCopyLink={handleCopyLink}
                    onCreateThread={handleCreateThread}
                    onBookmark={handleBookmark}
                    onOpenThread={handleOpenThread}
                    onEdit={handleEdit}
                    onDelete={handleDelete}
                  />
                ))}
                <div ref={messagesEndRef} />
              </div>
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center">
            <Text c="dimmed" size="sm">
              メッセージはまだありません
            </Text>
          </div>
        )}
      </div>

      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <form onSubmit={handleSubmit}>
          <div className="flex items-center gap-2">
            <FileInput onFileSelect={handleFileSelect} disabled={sendMessage.isPending || isUploading} />
            <div className="flex-1">
              <MessageInputToolbar
                mode={mode}
                onToggleMode={toggleMode}
                onSubmit={() => handleSubmit()}
                disabled={(body.trim().length === 0 && pendingAttachments.length === 0) || isUploading}
                loading={sendMessage.isPending}
                textareaRef={textareaRef}
              />
            </div>
          </div>
          {isEditMode ? (
            <Textarea
              ref={textareaRef}
              placeholder="メッセージを入力..."
              minRows={3}
              autosize
              value={body}
              onChange={(event) => handleBodyChange(event.currentTarget.value)}
              disabled={sendMessage.isPending}
            />
          ) : (
            <MessagePreview content={body} />
          )}

          {/* 添付ファイル一覧 */}
          <AttachmentList attachments={pendingAttachments} onRemove={removeAttachment} />

          {/* リンクプレビューを表示 */}
          {previews.length > 0 && (
            <div className="mt-3 space-y-2">
              {previews.map((preview) => (
                <LinkPreviewCard
                  key={preview.url}
                  preview={preview}
                  onRemove={() => removePreview(preview.url)}
                />
              ))}
            </div>
          )}

          {sendMessage.isError && (
            <Text c="red" size="sm" className="mt-2">
              {sendMessage.error?.message ?? "メッセージの送信に失敗しました"}
            </Text>
          )}
        </form>
      </Card>

      {/* スレッドサイドパネル */}
      {workspaceId && channelId && (
        <ThreadSidePanel
          opened={isThreadOpen}
          onClose={handleCloseThread}
          messageId={openThreadId}
          workspaceId={workspaceId}
          channelId={channelId}
        />
      )}
    </div>
  );
};
