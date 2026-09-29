import { useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useUserGroupActions } from "../hooks/useUserGroups";

import type { UserGroup } from "#/gen/chat/v1/user_group_service_pb";

type UserGroupDialogProps = {
  workspaceId: string;
  // null なら作成、あればそのグループを編集する
  group: UserGroup | null;
  // 保存したときはそのグループの ID を渡す
  onClose: (savedGroupId: string | null) => void;
};

export const UserGroupDialog = ({ workspaceId, group, onClose }: UserGroupDialogProps) => {
  const { t } = useTranslation();
  const { create, update } = useUserGroupActions();
  const [name, setName] = useState(group?.name ?? "");
  const [description, setDescription] = useState(group?.description ?? "");
  const failed = [create, update].find((mutation) => mutation.isError);

  const save = () => {
    const input = { description: description.trim(), name: name.trim() };
    if (group) {
      update.mutate(
        { ...input, groupId: group.id },
        {
          onSuccess: () => {
            toast(t("userGroup.updated", { name: input.name }));
            onClose(group.id);
          },
        },
      );
      return;
    }
    create.mutate(
      { ...input, workspaceId },
      {
        onSuccess: ({ userGroup }) => {
          toast(t("userGroup.created", { name: input.name }));
          onClose(userGroup?.id ?? null);
        },
      },
    );
  };

  return (
    <Dialog
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onClose(null);
        }
      }}
      title={t(group ? "userGroup.editTitle" : "userGroup.createLabel")}
      footer={
        <>
          <Button
            variant="secondary"
            onPress={() => {
              onClose(null);
            }}
          >
            {t("common.cancel")}
          </Button>
          <Button
            isDisabled={name.trim().length === 0}
            isPending={create.isPending || update.isPending}
            onPress={save}
          >
            {t(group ? "common.save" : "userGroup.create")}
          </Button>
        </>
      }
    >
      <TextField
        label={t("userGroup.name")}
        placeholder={t("userGroup.createPlaceholder")}
        description={t("userGroup.nameHint", { name: name.trim() || "group" })}
        value={name}
        onChange={setName}
        isRequired
      />
      <TextField label={t("userGroup.description")} value={description} onChange={setDescription} />
      {failed && <p className="m-0 text-caption text-danger">{failed.error.message}</p>}
    </Dialog>
  );
};
