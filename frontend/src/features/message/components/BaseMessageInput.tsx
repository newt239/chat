import { useState } from "react";
import type { ReactNode } from "react";

import { useMutation } from "@connectrpc/connect-query";
import { Form, TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { AttachmentListItem } from "#/features/attachment/components/AttachmentListItem";
import { useFileUpload } from "#/features/attachment/hooks/useFileUpload";
import { useDraftAutosave } from "#/features/draft/hooks/useDraftAutosave";
import { LinkPreviewCard } from "#/features/link/components/LinkPreviewCard";
import { LocationShareDialog } from "#/features/location/components/LocationShareDialog";
import { PendingLocation } from "#/features/location/components/PendingLocation";
import { useMentionCodec } from "#/features/mention/hooks/useMentionCodec";
import { PollComposerDialog } from "#/features/poll/components/PollComposerDialog";
import { VoiceRecorder } from "#/features/recorder/components/VoiceRecorder";
import { useScheduleMessage } from "#/features/schedule/hooks/useScheduledMessages";
import { CommandService } from "#/gen/chat/v1/command_service_pb";
import { MessageService } from "#/gen/chat/v1/message_service_pb";
import { useIsMobile } from "#/hooks/useMediaQuery";

import { useComposerTextarea } from "../hooks/useComposerTextarea";
import { useTypingNotifier } from "../hooks/useTypingNotifier";
import { findCommand, unescapeCommand } from "../utils/commands";
import { detectActiveFormats, insertEmoji, toggleFormat } from "../utils/format";
import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";
import { SuggestionList } from "./SuggestionList";

import type { Message, MessageLocation, PollInput } from "#/gen/chat/v1/message_pb";

type BaseMessageInputProps = {
  placeholder: string;
  channelId: string;
  // スレッドへの返信の欄のときの親メッセージ
  parentId: string | null;
  // 集約表示中の投稿先の切り替え。入力欄の上に出す
  targetPicker: ReactNode;
  // 送った投稿を呼び出し側の表示に足すとき
  onSent: ((message: Message) => void) | null;
};

const urlPattern = /https?:\/\/[^\s<>"{}|\\^`[\]]+/g;

export const BaseMessageInput = ({
  placeholder,
  channelId,
  parentId,
  targetPicker,
  onSent,
}: BaseMessageInputProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [typedBody, setTypedBody] = useState<string | null>(null);
  const [isPreview, setIsPreview] = useState(false);
  const [location, setLocation] = useState<MessageLocation | undefined>(undefined);
  const [isLocationOpen, setIsLocationOpen] = useState(false);
  const [isPollOpen, setIsPollOpen] = useState(false);
  const [isRecorderOpen, setIsRecorderOpen] = useState(false);
  const { notifyTyping, notifyStopTyping } = useTypingNotifier(channelId);
  const {
    discard: discardDraft,
    initialBody: draftBody,
    save: saveDraft,
  } = useDraftAutosave(channelId, parentId);
  const mentionCodec = useMentionCodec();
  const { decode, encode, isReady: isCodecReady } = mentionCodec;
  // 打ち始めるまでは書きかけの下書きを出す。名前に戻せるようメンバーを読んでから戻す
  const body = typedBody ?? (draftBody !== null && isCodecReady ? decode(draftBody) : "");
  // 本文の URL ごとにプレビューを出す。閉じたものは本文に残っていても出さない
  const [dismissedUrls, setDismissedUrls] = useState<string[]>([]);
  const previewUrls = [...new Set(body.match(urlPattern))].filter(
    (url) => !dismissedUrls.includes(url),
  );
  const {
    pendingAttachments,
    uploadFile,
    removeAttachment,
    clearAttachments,
    getCompletedAttachmentIds,
    isUploading,
  } = useFileUpload();
  const sendMessage = useMutation(MessageService.method.createMessage);
  const scheduleMessage = useScheduleMessage();
  // 応答は公式アプリの投稿として WebSocket で届く
  const executeCommand = useMutation(CommandService.method.executeCommand);
  const isBusy = sendMessage.isPending || executeCommand.isPending;

  const handleBodyChange = (next: string) => {
    setTypedBody(next);
    notifyTyping();
    saveDraft(encode(next));
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
    return {
      attachmentIds: getCompletedAttachmentIds(),
      body: encode(body.trim()),
      location,
      poll: undefined,
    };
  };

  // 送信中に入力欄を離れても送った本文が下書きに残らないよう、送る時点で消す
  const discardComposerDraft = () => {
    notifyStopTyping();
    discardDraft();
  };

  const clearComposer = () => {
    setTypedBody("");
    setLocation(undefined);
    setIsPreview(false);
    setDismissedUrls([]);
    clearAttachments();
  };

  const send = (content: {
    body: string;
    attachmentIds: string[];
    location: MessageLocation | undefined;
    poll: PollInput | undefined;
  }) => {
    discardComposerDraft();
    sendMessage.mutate(
      { ...content, channelId, parentId: parentId ?? undefined },
      {
        onError: () => {
          saveDraft(encode(body));
        },
        onSuccess: ({ message }) => {
          clearComposer();
          if (message && onSent) {
            onSent(message);
          }
        },
      },
    );
  };

  const handleSubmit = () => {
    const content = collectContent();
    if (content === null) {
      return;
    }
    // 添付や位置情報のない「/コマンド」は投稿せずに実行する
    if (
      findCommand(content.body) !== null &&
      content.attachmentIds.length === 0 &&
      content.location === undefined
    ) {
      discardComposerDraft();
      executeCommand.mutate(
        { channelId, parentId: parentId ?? undefined, text: content.body },
        {
          onError: (commandError) => {
            saveDraft(encode(body));
            toast(commandError.rawMessage || t("command.failed"), { tone: "danger" });
          },
          onSuccess: clearComposer,
        },
      );
      return;
    }
    send({ ...content, body: unescapeCommand(content.body) });
  };

  const composer = useComposerTextarea({
    allowsCommands: true,
    body,
    onBodyChange: handleBodyChange,
    onEscape: null,
    onSubmit: handleSubmit,
    registerMention: mentionCodec.register,
    // モバイルの Enter は改行にする
    submitsOnEnter: !isMobile,
  });
  const { selection, suggestion } = composer;

  const handleInsertEmoji = (emoji: string) => {
    const next = insertEmoji(body, composer.readSelection(), emoji);
    handleBodyChange(next.text);
    composer.setSelection({ end: next.cursor, start: next.cursor });
  };

  const handleSchedule = (scheduledAt: Date) => {
    const content = collectContent();
    if (content !== null) {
      discardComposerDraft();
      scheduleMessage(
        { ...content, channelId, parentId: parentId ?? undefined },
        scheduledAt,
        clearComposer,
      );
    }
  };

  return (
    <Form
      onSubmit={(event) => {
        event.preventDefault();
        handleSubmit();
      }}
      className="shrink-0 px-4.5 pb-3 font-sans max-md:px-2.5 max-md:pb-2"
    >
      {targetPicker}
      <div className="relative rounded-lg border border-border-strong bg-surface focus-within:border-accent focus-within:ring-3 focus-within:ring-accent-soft">
        {pendingAttachments.length > 0 && (
          <div className="flex flex-wrap gap-1.5 px-2.5 pt-2">
            {pendingAttachments.map((attachment) => (
              <AttachmentListItem
                key={attachment.id}
                attachment={attachment}
                onRemove={() => {
                  removeAttachment(attachment.id);
                }}
              />
            ))}
          </div>
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
        {!isPreview && suggestion.isOpen && <SuggestionList {...suggestion.listProps} />}
        {isPreview ? (
          <MessagePreview content={encode(body)} />
        ) : (
          <TextField aria-label={placeholder} {...composer.fieldProps} isDisabled={isBusy}>
            <TextArea
              {...composer.textAreaProps}
              rows={1}
              placeholder={placeholder}
              className="block max-h-45 min-h-9.5 w-full resize-none border-0 bg-transparent px-3 pt-2.25 pb-0.5 font-sans text-body leading-relaxed text-text outline-none [field-sizing:content] placeholder:text-subtle"
            />
          </TextField>
        )}
        {previewUrls.length > 0 && (
          <div className="flex flex-col gap-2 px-3 py-2">
            {previewUrls.map((url) => (
              <LinkPreviewCard
                key={url}
                url={url}
                onRemove={() => {
                  setDismissedUrls((prev) => [...prev, url]);
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
          isSendDisabled={isBusy || !hasContent || isUploading}
          isSending={isBusy}
          activeFormats={detectActiveFormats(body, selection)}
          onFormat={(key) => {
            composer.replaceSelection(toggleFormat(body, selection, key));
          }}
          onInsertEmoji={handleInsertEmoji}
          onFocusInput={() => {
            composer.focus();
          }}
          onFileSelect={(files) => {
            void handleFileSelect(files);
          }}
          onShareLocation={() => {
            setIsLocationOpen(true);
          }}
          onRecord={() => {
            setIsRecorderOpen(true);
          }}
          onCreatePoll={() => {
            setIsPollOpen(true);
          }}
          onSchedule={handleSchedule}
        />
      </div>
      {isPollOpen && (
        <PollComposerDialog
          onClose={() => {
            setIsPollOpen(false);
          }}
          onConfirm={(poll) => {
            // 書きかけの本文は投票の説明として一緒に投稿する
            send({ attachmentIds: [], body: encode(body.trim()), location: undefined, poll });
          }}
        />
      )}
      {isLocationOpen && (
        <LocationShareDialog
          onClose={() => {
            setIsLocationOpen(false);
          }}
          onConfirm={setLocation}
        />
      )}
      {sendMessage.isError && (
        <p className="m-0 mt-1.5 text-caption text-danger">
          {sendMessage.error.rawMessage || t("message.composer.sendFailed")}
        </p>
      )}
    </Form>
  );
};
