import { useRef, useState } from "react";
import type { ReactNode } from "react";

import { Form, TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { AttachmentListItem } from "#/features/attachment/components/AttachmentListItem";
import { useFileUpload } from "#/features/attachment/hooks/useFileUpload";
import { useExecuteCommand } from "#/features/command/hooks/useExecuteCommand";
import { findCommand, unescapeCommand } from "#/features/command/utils/commands";
import { useDraftAutosave } from "#/features/draft/hooks/useDraftAutosave";
import { LinkPreviewCard } from "#/features/link/components/LinkPreviewCard";
import { LocationShareDialog } from "#/features/location/components/LocationShareDialog";
import { PendingLocation } from "#/features/location/components/PendingLocation";
import { useMentionCodec } from "#/features/mention/hooks/useMentionCodec";
import { PollComposerDialog } from "#/features/poll/components/PollComposerDialog";
import { VoiceRecorder } from "#/features/recorder/components/VoiceRecorder";
import { useScheduleMessage } from "#/features/schedule/hooks/useScheduledMessages";
import { useIsMobile } from "#/hooks/useMediaQuery";

import { useComposerSuggestion } from "../hooks/useComposerSuggestion";
import { useTypingNotifier } from "../hooks/useTypingNotifier";
import { continueList, detectActiveFormats, insertEmoji, toggleFormat } from "../utils/format";
import { MessageInputToolbar } from "./MessageInputToolbar";
import { MessagePreview } from "./MessagePreview";
import { SuggestionList } from "./SuggestionList";

import type { FormatKey, Selection } from "../utils/format";

import type { MessageLocation, PollInput } from "#/gen/chat/v1/message_pb";

// 入力欄から送る内容。CreateMessage の入力にそのまま広げて使う
export type ComposerContent = {
  body: string;
  attachmentIds: string[];
  location: MessageLocation | undefined;
  poll: PollInput | undefined;
};

type BaseMessageInputProps = {
  onSubmit: (content: ComposerContent) => void;
  placeholder: string;
  isPending: boolean;
  error: string | null;
  channelId: string;
  // スレッドへの返信の欄のときの親メッセージ。下書きの置き場所に使う
  parentId: string | null;
  // 集約表示中の投稿先の切り替え。入力欄の上に出す
  targetPicker: ReactNode;
};

const urlPattern = /https?:\/\/[^\s<>"{}|\\^`[\]]+/g;

export const BaseMessageInput = ({
  onSubmit,
  placeholder,
  isPending,
  error,
  channelId,
  parentId,
  targetPicker,
}: BaseMessageInputProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [typedBody, setTypedBody] = useState<string | null>(null);
  const [selection, setSelection] = useState({ end: 0, start: 0 });
  const [isPreview, setIsPreview] = useState(false);
  const [location, setLocation] = useState<MessageLocation | undefined>(undefined);
  const [isLocationOpen, setIsLocationOpen] = useState(false);
  const [isPollOpen, setIsPollOpen] = useState(false);
  const [isRecorderOpen, setIsRecorderOpen] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
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
  const scheduleMessage = useScheduleMessage();
  const executeCommand = useExecuteCommand();
  const isBusy = isPending || executeCommand.isPending;

  const handleBodyChange = (next: string) => {
    setTypedBody(next);
    notifyTyping();
    saveDraft(encode(next));
  };

  const replaceSelection = (next: { text: string; selection: Selection }) => {
    handleBodyChange(next.text);
    setSelection(next.selection);
    // 値が反映されてから選択範囲を動かす
    requestAnimationFrame(() => {
      textareaRef.current?.focus();
      textareaRef.current?.setSelectionRange(next.selection.start, next.selection.end);
    });
  };

  const suggestion = useComposerSuggestion({
    allowsCommands: true,
    body,
    cursor: selection.start,
    onApply: (next, item) => {
      mentionCodec.register(item.value, item.token);
      replaceSelection({ selection: { end: next.cursor, start: next.cursor }, text: next.text });
    },
  });

  const handleFormat = (key: FormatKey) => {
    replaceSelection(toggleFormat(body, selection, key));
  };

  const handleInsertEmoji = (emoji: string) => {
    // select イベントはカーソルの移動では届かないため、挿入時の位置を textarea から読む
    const textarea = textareaRef.current;
    const current =
      textarea === null
        ? selection
        : { end: textarea.selectionEnd, start: textarea.selectionStart };
    const next = insertEmoji(body, current, emoji);
    handleBodyChange(next.text);
    setSelection({ end: next.cursor, start: next.cursor });
  };

  const focusInput = () => {
    textareaRef.current?.focus();
    textareaRef.current?.setSelectionRange(selection.start, selection.end);
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

  // 送信・予約した後は書きかけも消す
  const resetComposer = () => {
    notifyStopTyping();
    discardDraft();
    setTypedBody("");
    setLocation(undefined);
    setIsPreview(false);
    setDismissedUrls([]);
    clearAttachments();
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
      executeCommand.mutate(
        { channelId, parentId: parentId ?? undefined, text: content.body },
        {
          onError: (commandError) => {
            toast(commandError.rawMessage || t("command.failed"), { tone: "danger" });
          },
          onSuccess: resetComposer,
        },
      );
      return;
    }
    onSubmit({ ...content, body: unescapeCommand(content.body) });
    resetComposer();
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
          <TextField
            aria-label={placeholder}
            value={body}
            onChange={(next) => {
              handleBodyChange(next);
              // 候補の検索語はカーソルの位置で決まるため、入力のたびに読み直す
              const textarea = textareaRef.current;
              if (textarea !== null) {
                setSelection({ end: textarea.selectionEnd, start: textarea.selectionStart });
              }
            }}
            isDisabled={isBusy}
            onKeyDown={(event) => {
              if (
                suggestion.handleKeyDown(event) ||
                event.key !== "Enter" ||
                event.nativeEvent.isComposing
              ) {
                return;
              }
              // モバイルの Enter は改行にする
              if (!event.shiftKey && !isMobile) {
                event.preventDefault();
                handleSubmit();
                return;
              }
              const textarea = textareaRef.current;
              const continued =
                textarea &&
                continueList(body, { end: textarea.selectionEnd, start: textarea.selectionStart });
              if (continued) {
                event.preventDefault();
                replaceSelection(continued);
              }
            }}
          >
            <TextArea
              {...suggestion.inputProps}
              ref={textareaRef}
              rows={1}
              placeholder={placeholder}
              onSelect={(event) => {
                setSelection({
                  end: event.currentTarget.selectionEnd,
                  start: event.currentTarget.selectionStart,
                });
              }}
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
          onFormat={handleFormat}
          onInsertEmoji={handleInsertEmoji}
          onFocusInput={focusInput}
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
      <PollComposerDialog
        isOpen={isPollOpen}
        onOpenChange={setIsPollOpen}
        onConfirm={(poll) => {
          // 書きかけの本文は投票の説明として一緒に投稿する
          onSubmit({ attachmentIds: [], body: encode(body.trim()), location: undefined, poll });
          resetComposer();
        }}
      />
      <LocationShareDialog
        isOpen={isLocationOpen}
        onOpenChange={setIsLocationOpen}
        onConfirm={setLocation}
      />
      {error && <p className="m-0 mt-1.5 text-caption text-danger">{error}</p>}
    </Form>
  );
};
