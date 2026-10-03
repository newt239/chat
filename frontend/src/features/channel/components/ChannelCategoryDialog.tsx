import { useId, useState } from "react";

import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useChannelCategoryActions } from "../hooks/useChannelCategories";

import type { ChannelCategory } from "#/gen/chat/v1/channel_category_service_pb";

type ChannelCategoryDialogProps = {
  workspaceId: string;
  // null なら新しく作る
  category: ChannelCategory | null;
  // 作成後にそのカテゴリへ移すチャンネル
  assignChannelId: string | null;
  onClose: () => void;
};

export const ChannelCategoryDialog = ({
  workspaceId,
  category,
  assignChannelId,
  onClose,
}: ChannelCategoryDialogProps) => {
  const { t } = useTranslation();
  const formId = useId();
  const [name, setName] = useState(category?.name ?? "");
  const { create, setChannel, update } = useChannelCategoryActions(workspaceId);
  const trimmed = name.trim();

  const submit = () => {
    if (trimmed === "") {
      return;
    }
    if (category !== null) {
      update.mutate({ categoryId: category.id, name: trimmed }, { onSuccess: onClose });
      return;
    }
    create.mutate(
      { name: trimmed, workspaceId },
      {
        onSuccess: ({ category: created }) => {
          toast(t("channel.category.created", { name: trimmed }), { tone: "success" });
          if (created !== undefined && assignChannelId !== null) {
            setChannel.mutate({ categoryId: created.id, channelId: assignChannelId });
          }
          onClose();
        },
      },
    );
  };

  return (
    <Dialog
      isOpen
      onOpenChange={onClose}
      title={category === null ? t("channel.category.create") : t("channel.category.renameTitle")}
      footer={
        <>
          <Button variant="secondary" onPress={onClose}>
            {t("common.cancel")}
          </Button>
          <Button
            type="submit"
            form={formId}
            isDisabled={trimmed === ""}
            isPending={create.isPending || update.isPending}
          >
            {category === null ? t("channel.create.submit") : t("common.save")}
          </Button>
        </>
      }
    >
      <Form
        id={formId}
        onSubmit={(event) => {
          event.preventDefault();
          submit();
        }}
      >
        <TextField
          label={t("channel.category.name")}
          value={name}
          onChange={setName}
          maxLength={50}
        />
      </Form>
    </Dialog>
  );
};
