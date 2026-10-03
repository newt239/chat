import { useState } from "react";

import { TextArea, TextField } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { useMentionCodec } from "#/features/mention/hooks/useMentionCodec";

import { useComposerTextarea } from "../hooks/useComposerTextarea";
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
  const [editedDraft, setEditedDraft] = useState<string | null>(null);
  // 編集を始めるまでは本文の ID 記法を名前に戻して出す。メンバーを読み込む前に開いたら、読み込んだときに戻る
  const draft = editedDraft ?? mentionCodec.decode(initialBody);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
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
  const composer = useComposerTextarea({
    allowsCommands: false,
    body: draft,
    onBodyChange: setEditedDraft,
    onEscape: onClose,
    onSubmit: () => {
      void save();
    },
    registerMention: mentionCodec.register,
    submitsOnEnter: true,
  });

  return (
    <div className="relative flex flex-col gap-1.5">
      {composer.suggestion.isOpen && <SuggestionList {...composer.suggestion.listProps} />}
      <TextField
        aria-label={t("message.actions.edit")}
        {...composer.fieldProps}
        isDisabled={isSaving}
        isInvalid={error !== null}
        // oxlint-disable-next-line jsx-a11y/no-autofocus -- 編集を始めた直後に入力できるようにする
        autoFocus
      >
        <TextArea
          {...composer.textAreaProps}
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
