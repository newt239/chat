import { useState } from "react";

import { IconLock } from "@tabler/icons-react";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { TextArea } from "#/components/ui/TextArea/TextArea";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useUpdateUserNote } from "../hooks/useUserNote";

type UserNoteEditorProps = {
  targetUserId: string;
  // 未設定なら表示名をプレースホルダーに出す
  displayName: string;
  initialNickname: string;
  initialMemo: string;
};

// 自分だけに見えるニックネームとメモ。両方を空にして保存すると削除される
export const UserNoteEditor = ({
  targetUserId,
  displayName,
  initialNickname,
  initialMemo,
}: UserNoteEditorProps) => {
  const { t } = useTranslation();
  const [nickname, setNickname] = useState(initialNickname);
  const [memo, setMemo] = useState(initialMemo);
  const update = useUpdateUserNote();
  const isDirty = nickname !== initialNickname || memo !== initialMemo;

  return (
    <Form
      className="flex flex-col gap-2.5 border-b border-border px-4 py-3"
      onSubmit={(event) => {
        event.preventDefault();
        update.mutate(
          { memo: memo.trim(), nickname: nickname.trim(), targetUserId },
          {
            onSuccess: () => {
              toast(t("member.note.saved"));
            },
          },
        );
      }}
    >
      <h4 className="m-0 flex items-center gap-1 text-xs font-semibold text-muted [&_svg]:size-3.5">
        <IconLock aria-hidden />
        {t("member.note.title")}
      </h4>
      <TextField
        label={t("member.note.nickname")}
        placeholder={displayName}
        value={nickname}
        onChange={setNickname}
        maxLength={50}
      />
      <TextArea
        label={t("member.note.memo")}
        placeholder={t("member.note.memoPlaceholder")}
        value={memo}
        onChange={setMemo}
        rows={4}
        maxLength={2000}
      />
      <div className="flex justify-end">
        <Button type="submit" size="sm" isDisabled={!isDirty} isPending={update.isPending}>
          {t("common.save")}
        </Button>
      </div>
      {update.isError && <p className="m-0 text-caption text-danger">{update.error.message}</p>}
    </Form>
  );
};
