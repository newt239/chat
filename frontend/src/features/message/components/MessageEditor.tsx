import { useEffect, useRef, useState } from "react";

import { TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";

import { useComposerSuggestion } from "../hooks/useComposerSuggestion";
import { useMentionCodec } from "../hooks/useMentionCodec";
import { SuggestionList } from "./SuggestionList";

type MessageEditorProps = {
  initialBody: string;
  // 失敗したら reject する。エラーの通知は呼び出し側が行う
  onSave: (body: string) => Promise<void>;
  onClose: () => void;
};

export const MessageEditor = ({ initialBody, onSave, onClose }: MessageEditorProps) => {
  const { t } = useTranslation();
  const mentionCodec = useMentionCodec();
  // 本文の ID 記法を名前に戻して編集させる。メンバーを読み込む前に開いたら、読み込んだときに戻す
  const [draft, setDraft] = useState(() => mentionCodec.decode(initialBody));
  const { decode, isReady } = mentionCodec;
  useEffect(() => {
    if (isReady) {
      setDraft((current) => (current === initialBody ? decode(initialBody) : current));
    }
  }, [isReady, decode, initialBody]);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [cursor, setCursor] = useState(draft.length);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const suggestion = useComposerSuggestion({
    allowsCommands: false,
    body: draft,
    cursor,
    onApply: (next, item) => {
      mentionCodec.register(item.value, item.token);
      setDraft(next.text);
      setCursor(next.cursor);
      requestAnimationFrame(() => {
        textareaRef.current?.setSelectionRange(next.cursor, next.cursor);
      });
    },
  });
  const syncCursor = () => {
    setCursor(textareaRef.current?.selectionStart ?? 0);
  };

  const save = async () => {
    const trimmed = mentionCodec.encode(draft.trim());
    if (trimmed.length === 0) {
      setError(t("message.edit.empty"));
      return;
    }
    if (trimmed === initialBody) {
      onClose();
      return;
    }
    setIsSaving(true);
    try {
      await onSave(trimmed);
      onClose();
    } catch {
      setIsSaving(false);
    }
  };

  return (
    <div className="relative flex flex-col gap-1.5">
      {suggestion.isOpen && <SuggestionList {...suggestion.listProps} />}
      <TextField
        aria-label={t("message.actions.edit")}
        value={draft}
        onChange={(next) => {
          setDraft(next);
          syncCursor();
        }}
        isDisabled={isSaving}
        isInvalid={error !== null}
        // oxlint-disable-next-line jsx-a11y/no-autofocus -- 編集を始めた直後に入力できるようにする
        autoFocus
        onKeyDown={(event) => {
          if (suggestion.handleKeyDown(event)) {
            return;
          }
          if (event.key === "Escape") {
            onClose();
          }
          if (event.key === "Enter" && !event.shiftKey && !event.nativeEvent.isComposing) {
            event.preventDefault();
            void save();
          }
        }}
      >
        <TextArea
          {...suggestion.inputProps}
          ref={textareaRef}
          onSelect={syncCursor}
          className="min-h-[60px] w-full resize-y rounded-md border border-accent bg-surface px-2.5 py-1.5 font-sans text-body text-text ring-3 ring-accent-soft outline-none"
        />
      </TextField>
      {error && <p className="m-0 text-caption text-danger">{error}</p>}
      <div className="flex items-center justify-end gap-1.5 text-[11.5px] text-subtle">
        <span className="flex-1">{t("message.edit.hint")}</span>
        <Button variant="secondary" size="sm" onPress={onClose} isDisabled={isSaving}>
          {t("common.cancel")}
        </Button>
        <Button
          size="sm"
          isPending={isSaving}
          onPress={() => {
            void save();
          }}
        >
          {t("common.save")}
        </Button>
      </div>
    </div>
  );
};
