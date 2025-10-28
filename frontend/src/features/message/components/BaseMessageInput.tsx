import { useState, useRef, useCallback, useEffect } from "react";
import type { FormEvent } from "react";

import { Text, Textarea } from "@mantine/core";
import { notifications } from "@mantine/notifications";

import { useMessageInputMode } from "../hooks/useMessageInputMode";

import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";

import { AttachmentList } from "@/features/attachment/components/AttachmentList";
import { useFileUpload } from "@/features/attachment/hooks/useFileUpload";
import { LinkPreviewCard } from "@/features/link/components/LinkPreviewCard";
import { useLinkPreview } from "@/features/link/hooks/useLinkPreview";

type BaseMessageInputProps = {
  onSubmit: (body: string, attachmentIds: string[]) => void;
  placeholder?: string;
  isPending?: boolean;
  error?: string;
  channelId?: string | null;
  onReset?: () => void;
};

export const BaseMessageInput = ({
  onSubmit,
  placeholder = "メッセージを入力...",
  isPending = false,
  error,
  channelId,
  onReset,
}: BaseMessageInputProps) => {
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

  const handleBodyChange = useCallback(
    (newValue: string) => {
      setBody(newValue);

      // URLを検出してプレビューを追加
      const urlRegex = /https?:\/\/[^\s<>"{}|\\^`\[\]]+/g;
      const urls: string[] = newValue.match(urlRegex) || [];

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
    onSubmit(body.trim(), attachmentIds);
    setBody("");
    clearPreviews();
    clearAttachments();
  };

  // 外部からリセットが呼ばれた場合
  useEffect(() => {
    if (onReset) {
      setBody("");
      clearPreviews();
      clearAttachments();
    }
  }, [onReset, clearPreviews, clearAttachments]);

  const isDisabled =
    isPending || (body.trim().length === 0 && pendingAttachments.length === 0) || isUploading;

  return (
    <form
      onSubmit={handleSubmit}
      className="p-2 border-y border-gray-200"
      style={{ backgroundColor: isEditMode ? "gray.100" : "white" }}
    >
      {isEditMode ? (
        <Textarea
          ref={textareaRef}
          placeholder={placeholder}
          minRows={3}
          autosize
          value={body}
          onChange={(event) => handleBodyChange(event.currentTarget.value)}
          disabled={isPending}
        />
      ) : (
        <MessagePreview content={body} />
      )}

      {/* 添付ファイル一覧 */}
      {pendingAttachments.length > 0 && (
        <AttachmentList attachments={pendingAttachments} onRemove={removeAttachment} />
      )}

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

      <MessageInputToolbar
        mode={mode}
        onToggleMode={toggleMode}
        onSubmit={() => handleSubmit()}
        disabled={isDisabled}
        loading={isPending}
        textareaRef={textareaRef}
        onFileSelect={handleFileSelect}
        isFileUploadDisabled={isPending || isUploading}
      />

      {error && (
        <Text c="red" size="sm" className="mt-2">
          {error}
        </Text>
      )}
    </form>
  );
};
