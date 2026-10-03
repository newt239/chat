import { useRef, useState } from "react";

import { TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { useMentionCodec } from "#/features/mention/hooks/useMentionCodec";

import { useComposerSuggestion } from "../hooks/useComposerSuggestion";
import { handleEnterKey } from "../utils/format";
import { SuggestionList } from "./SuggestionList";

import type { Selection } from "../utils/format";

type MessageEditorProps = {
  initialBody: string;
  // 失敗したら reject する。エラーの通知は呼び出し側が行う
  onSave: (body: string) => Promise<void>;
  onClose: () => void;
};

export const MessageEditor = ({ initialBody, onSave, onClose }: MessageEditorProps) => {
  const { t } = useTranslation();
  const mentionCodec = useMentionCodec();
  const [editedDraft, setEditedDraft] = useState<string | null>(null);
  // 編集を始めるまでは本文の ID 記法を名前に戻して出す。メンバーを読み込む前に開いたら、読み込んだときに戻る
  const draft = editedDraft ?? mentionCodec.decode(initialBody);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selection, setSelection] = useState({ end: draft.length, start: draft.length });
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const replaceSelection = (next: { text: string; selection: Selection }) => {
    setEditedDraft(next.text);
    setSelection(next.selection);
    requestAnimationFrame(() => {
      textareaRef.current?.setSelectionRange(next.selection.start, next.selection.end);
    });
  };
  const suggestion = useComposerSuggestion({
    allowsCommands: false,
    body: draft,
    cursor: selection.start,
    onApply: (next, item) => {
      mentionCodec.register(item.value, item.token);
      replaceSelection({ selection: { end: next.cursor, start: next.cursor }, text: next.text });
    },
  });
  const syncSelection = () => {
    const textarea = textareaRef.current;
    if (textarea !== null) {
      setSelection({ end: textarea.selectionEnd, start: textarea.selectionStart });
    }
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
          setEditedDraft(next);
          syncSelection();
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
          handleEnterKey({
            event,
            onReplace: replaceSelection,
            onSubmit: () => {
              void save();
            },
            submitsOnEnter: true,
            text: draft,
            textarea: textareaRef.current,
          });
        }}
      >
        <TextArea
          {...suggestion.inputProps}
          ref={textareaRef}
          onSelect={syncSelection}
          className="min-h-15 w-full resize-y rounded-md border border-accent bg-surface px-2.5 py-1.5 font-sans text-body text-text ring-3 ring-accent-soft outline-none"
        />
      </TextField>
      {error && <p className="m-0 text-caption text-danger">{error}</p>}
      <div className="flex items-center justify-end gap-1.5 text-caption text-subtle">
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
