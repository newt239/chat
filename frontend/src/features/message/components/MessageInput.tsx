import { useState, useRef, useCallback, useEffect } from "react";
import type { FormEvent } from "react";

import { Card, Text, Textarea } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useAtomValue } from "jotai";

import { useSendMessage } from "../hooks/useMessage";
import { useMessageInputMode } from "../hooks/useMessageInputMode";

import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";

import { AttachmentList } from "@/features/attachment/components/AttachmentList";
import { FileInput } from "@/features/attachment/components/FileInput";
import { useFileUpload } from "@/features/attachment/hooks/useFileUpload";
import { LinkPreviewCard } from "@/features/link/components/LinkPreviewCard";
import { useLinkPreview } from "@/features/link/hooks/useLinkPreview";
import { currentChannelIdAtom } from "@/providers/store/workspace";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const sendMessage = useSendMessage(channelId);
  const [body, setBody] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
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

  const handleFileSelect = useCallback(
    async (files: File[]) => {
      if (!channelId) return;

      for (const file of files) {
        await uploadFile(file, { channelId });
      }
    },
    [channelId, uploadFile]
  );

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

  // チャンネルが切り替わった時にフォームをリセット
  const resetForm = useCallback(() => {
    setBody("");
    clearPreviews();
    clearAttachments();
  }, [clearPreviews, clearAttachments]);

  // チャンネルが変更された時にフォームをリセット
  useEffect(() => {
    if (channelId !== currentChannelId) {
      resetForm();
    }
  }, [channelId, currentChannelId, resetForm]);

  if (!channelId) {
    return null;
  }

  return (
    <Card withBorder padding="lg" radius="md" className="shrink-0">
      <form onSubmit={handleSubmit}>
        <div className="flex items-center gap-2">
          <FileInput
            onFileSelect={handleFileSelect}
            disabled={sendMessage.isPending || isUploading}
          />
          <div className="flex-1">
            <MessageInputToolbar
              mode={mode}
              onToggleMode={toggleMode}
              onSubmit={() => handleSubmit()}
              disabled={
                (body.trim().length === 0 && pendingAttachments.length === 0) || isUploading
              }
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
  );
};
