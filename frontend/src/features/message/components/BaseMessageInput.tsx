import { useCallback, useEffect, useRef, useState } from "react";
import type { ReactNode } from "react";

import { Form, TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import { AttachmentList } from "#/features/attachment/components/AttachmentList";
import { useFileUpload } from "#/features/attachment/hooks/useFileUpload";
import { useDraftAutosave } from "#/features/draft/hooks/useDraftAutosave";
import { LinkPreviewCard } from "#/features/link/components/LinkPreviewCard";
import { useLinkPreview } from "#/features/link/hooks/useLinkPreview";
import { LocationShareDialog } from "#/features/location/components/LocationShareDialog";
import { PendingLocation } from "#/features/location/components/PendingLocation";
import { VoiceRecorder } from "#/features/recorder/components/VoiceRecorder";
import { useScheduleMessage } from "#/features/schedule/hooks/useScheduledMessages";
import { useIsMobile } from "#/lib/useMediaQuery";

import { useTypingNotifier } from "../hooks/useTypingNotifier";
import { applyFormat, detectActiveFormats } from "../utils/format";
import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";

import type { ComposerContent } from "../utils/composerContent";
import type { FormatKey } from "../utils/format";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type BaseMessageInputProps = {
  onSubmit: (content: ComposerContent) => void;
  placeholder: string;
  isPending: boolean;
  error?: string;
  channelId: string;
  // スレッドへの返信の欄のときの親メッセージ。下書きの置き場所に使う
  parentId: string | null;
  // 集約表示中の投稿先の切り替え。入力欄の上に出す
  targetPicker?: ReactNode;
};

const urlPattern = /https?:\/\/[^\s<>"{}|\\^`[\]]+/g;

export const BaseMessageInput = ({
  onSubmit,
  placeholder,
  isPending,
  error,
  channelId,
  parentId,
  targetPicker = null,
}: BaseMessageInputProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [body, setBody] = useState("");
  const [selection, setSelection] = useState({ end: 0, start: 0 });
  const [isPreview, setIsPreview] = useState(false);
  const [location, setLocation] = useState<MessageLocation | undefined>(undefined);
  const [isLocationOpen, setIsLocationOpen] = useState(false);
  const [isRecorderOpen, setIsRecorderOpen] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { previews, addPreview, removePreview, clearPreviews } = useLinkPreview();
  const {
    pendingAttachments,
    uploadFile,
    removeAttachment,
    clearAttachments,
    getCompletedAttachmentIds,
    isUploading,
  } = useFileUpload();
  const { notifyTyping, notifyStopTyping } = useTypingNotifier(channelId);
  const {
    discard: discardDraft,
    initialBody: draftBody,
    save: saveDraft,
  } = useDraftAutosave(channelId, parentId);
  const isRestoredRef = useRef(false);
  const scheduleMessage = useScheduleMessage();

  // 開き直したときに書きかけを戻す。読み込み前に打ち始めていたら上書きしない
  useEffect(() => {
    if (draftBody === null || isRestoredRef.current) {
      return;
    }
    isRestoredRef.current = true;
    setBody((current) => (current === "" ? draftBody : current));
  }, [draftBody]);

  const handleBodyChange = useCallback(
    (next: string) => {
      setBody(next);
      notifyTyping();
      saveDraft(next);

      const urls: string[] = next.match(urlPattern) ?? [];
      for (const url of urls) {
        void addPreview(url);
      }
      for (const preview of previews) {
        if (!urls.includes(preview.url)) {
          removePreview(preview.url);
        }
      }
    },
    [addPreview, previews, removePreview, notifyTyping, saveDraft],
  );

  const handleFormat = (key: FormatKey) => {
    const next = applyFormat(body, selection, key);
    handleBodyChange(next.text);
    setSelection({ end: next.cursor, start: next.cursor });
    // 値が反映されてからカーソルを動かす
    requestAnimationFrame(() => {
      textareaRef.current?.focus();
      textareaRef.current?.setSelectionRange(next.cursor, next.cursor);
    });
  };

  const handleFileSelect = async (files: File[]) => {
    for (const file of files) {
      await uploadFile(file, { channelId, durationSeconds: undefined });
    }
  };

  const hasContent =
    body.trim().length > 0 || pendingAttachments.length > 0 || location !== undefined;

  const collectContent = () => {
    if (!hasContent) {
      return null;
    }
    if (isUploading) {
      toast(t("message.composer.uploading"));
      return null;
    }
    return { attachmentIds: getCompletedAttachmentIds(), body: body.trim(), location };
  };

  // 送信・予約した後は書きかけも消す
  const resetComposer = () => {
    notifyStopTyping();
    discardDraft();
    setBody("");
    setLocation(undefined);
    setIsPreview(false);
    clearPreviews();
    clearAttachments();
  };

  const handleSubmit = () => {
    const content = collectContent();
    if (content !== null) {
      onSubmit(content);
      resetComposer();
    }
  };

  const handleSchedule = (scheduledAt: Date) => {
    const content = collectContent();
    if (content !== null) {
      scheduleMessage(
        { ...content, channelId, parentId: parentId ?? undefined },
        scheduledAt,
        resetComposer,
      );
    }
  };

  return (
    <Form
      onSubmit={(event) => {
        event.preventDefault();
        handleSubmit();
      }}
      className="shrink-0 px-[18px] pb-3 font-sans max-md:px-2.5 max-md:pb-2"
    >
      {targetPicker}
      <div className="rounded-lg border border-border-strong bg-surface focus-within:border-accent focus-within:ring-3 focus-within:ring-accent-soft">
        {pendingAttachments.length > 0 && (
          <AttachmentList attachments={pendingAttachments} onRemove={removeAttachment} />
        )}
        {isRecorderOpen && (
          <VoiceRecorder
            onAttach={(file, durationSeconds) => {
              setIsRecorderOpen(false);
              void uploadFile(file, { channelId, durationSeconds });
            }}
            onDiscard={() => {
              setIsRecorderOpen(false);
            }}
          />
        )}
        {location && (
          <PendingLocation
            location={location}
            onRemove={() => {
              setLocation(undefined);
            }}
          />
        )}
        {isPreview ? (
          <MessagePreview content={body} />
        ) : (
          <TextField
            aria-label={placeholder}
            value={body}
            onChange={handleBodyChange}
            isDisabled={isPending}
            onKeyDown={(event) => {
              // モバイルの Enter は改行にする
              if (
                event.key === "Enter" &&
                !event.shiftKey &&
                !event.nativeEvent.isComposing &&
                !isMobile
              ) {
                event.preventDefault();
                handleSubmit();
              }
            }}
          >
            <TextArea
              ref={textareaRef}
              rows={1}
              placeholder={placeholder}
              onSelect={(event) => {
                setSelection({
                  end: event.currentTarget.selectionEnd,
                  start: event.currentTarget.selectionStart,
                });
              }}
              className="block max-h-[180px] min-h-[38px] w-full resize-none border-0 bg-transparent px-3 pt-[9px] pb-0.5 font-sans text-body leading-[1.6] text-text outline-none [field-sizing:content] placeholder:text-subtle"
            />
          </TextField>
        )}
        {previews.length > 0 && (
          <div className="flex flex-col gap-2 px-3 py-2">
            {previews.map((preview) => (
              <LinkPreviewCard
                key={preview.url}
                preview={preview}
                onRemove={() => {
                  removePreview(preview.url);
                }}
              />
            ))}
          </div>
        )}
        <MessageInputToolbar
          isPreview={isPreview}
          onTogglePreview={() => {
            setIsPreview((current) => !current);
          }}
          onSubmit={handleSubmit}
          isSendDisabled={isPending || !hasContent || isUploading}
          isSending={isPending}
          activeFormats={detectActiveFormats(body, selection)}
          onFormat={handleFormat}
          onFileSelect={(files) => {
            void handleFileSelect(files);
          }}
          onShareLocation={() => {
            setIsLocationOpen(true);
          }}
          onRecord={() => {
            setIsRecorderOpen(true);
          }}
          onSchedule={handleSchedule}
        />
      </div>
      <LocationShareDialog
        isOpen={isLocationOpen}
        onOpenChange={setIsLocationOpen}
        onConfirm={setLocation}
      />
      {error && <p className="m-0 mt-1.5 text-caption text-danger">{error}</p>}
    </Form>
  );
};
