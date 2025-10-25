import { useState, useRef, useCallback } from "react";
import type { FormEvent } from "react";

import { Text, Textarea } from "@mantine/core";

import { useMessageInputMode } from "../hooks/useMessageInputMode";

import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";

import { LinkPreviewCard } from "@/features/link/components/LinkPreviewCard";
import { useLinkPreview } from "@/features/link/hooks/useLinkPreview";

type ThreadReplyInputProps = {
  onSubmit: (body: string) => void;
  isPending: boolean;
  isError: boolean;
  errorMessage?: string;
};

export const ThreadReplyInput = ({
  onSubmit,
  isPending,
  isError,
  errorMessage,
}: ThreadReplyInputProps) => {
  const [body, setBody] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { mode, toggleMode, isEditMode } = useMessageInputMode();
  const linkPreview = useLinkPreview();
  const { previews, addPreview, removePreview, clearPreviews } = linkPreview;

  const handleBodyChange = useCallback(
    (value: string) => {
      setBody(value);

      const urlRegex = /https?:\/\/[^\s<>"{}|\\^`\[\]]+/g;
      const urls: string[] = value.match(urlRegex) || [];

      urls.forEach((url: string) => {
        if (!previews.some((preview) => preview.url === url)) {
          addPreview(url);
        }
      });

      previews.forEach((preview) => {
        const previewUrl: string = preview.url;
        if (!urls.includes(previewUrl)) {
          removePreview(previewUrl);
        }
      });
    },
    [previews, addPreview, removePreview]
  );

  const handleSubmit = (event?: FormEvent<HTMLFormElement>) => {
    event?.preventDefault();
    if (body.trim().length === 0) {
      return;
    }
    onSubmit(body.trim());
    setBody("");
    clearPreviews();
  };

  return (
    <div className="border-t pt-4">
      <form onSubmit={handleSubmit}>
        <MessageInputToolbar
          mode={mode}
          onToggleMode={toggleMode}
          onSubmit={() => handleSubmit()}
          disabled={body.trim().length === 0}
          loading={isPending}
          textareaRef={textareaRef}
        />
        {isEditMode ? (
          <Textarea
            ref={textareaRef}
            placeholder="スレッドに返信..."
            minRows={3}
            autosize
            value={body}
            onChange={(event) => handleBodyChange(event.currentTarget.value)}
            disabled={isPending}
          />
        ) : (
          <MessagePreview content={body} />
        )}

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

        {isError && (
          <Text c="red" size="sm" className="mt-2">
            {errorMessage ?? "返信の送信に失敗しました"}
          </Text>
        )}
      </form>
    </div>
  );
};
