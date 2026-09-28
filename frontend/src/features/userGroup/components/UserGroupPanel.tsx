import { useState } from "react";

import { IconAt, IconEdit, IconTrash } from "@tabler/icons-react";
import { useSetAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Button } from "#/components/ui/Button";
import { toast } from "#/components/ui/toast";
import { closeRightSidePanelAtom } from "#/providers/store/ui";

import { useUserGroupActions, useUserGroups } from "../hooks/useUserGroups";
import { UserGroupDialog } from "./UserGroupDialog";
import { UserGroupMembers } from "./UserGroupMembers";

type UserGroupPanelProps = {
  workspaceId: string;
  groupId: string;
};

// 右パネル（モバイルでは全画面）に出すユーザーグループの詳細と編集
export const UserGroupPanel = ({ workspaceId, groupId }: UserGroupPanelProps) => {
  const { t } = useTranslation();
  const { data: groups, isLoading } = useUserGroups(workspaceId);
  const { remove } = useUserGroupActions();
  const closePanel = useSetAtom(closeRightSidePanelAtom);
  const [dialog, setDialog] = useState<"edit" | "delete" | null>(null);
  const group = groups?.find((candidate) => candidate.id === groupId);

  if (group === undefined) {
    return isLoading ? null : (
      <p className="m-0 p-4 text-caption text-muted">{t("userGroup.notFound")}</p>
    );
  }

  return (
    <div className="flex min-h-full flex-col bg-surface font-sans text-text">
      <section className="flex flex-col gap-2 border-b border-border px-4 pt-4 pb-3.5">
        <h3 className="m-0 text-[19px] font-bold text-accent-text">@{group.name}</h3>
        {group.description !== undefined && group.description.length > 0 && (
          <p className="m-0 text-[13px]">{group.description}</p>
        )}
        <div className="flex flex-wrap gap-1.5">
          <Button
            variant="secondary"
            size="sm"
            onPress={() => {
              setDialog("edit");
            }}
          >
            <IconEdit aria-hidden />
            {t("userGroup.edit")}
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onPress={() => {
              void navigator.clipboard.writeText(`@${group.name}`);
              toast(t("userGroup.mentionCopied"));
            }}
          >
            <IconAt aria-hidden />
            {t("userGroup.copyMention")}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="text-danger"
            onPress={() => {
              setDialog("delete");
            }}
          >
            <IconTrash aria-hidden />
            {t("common.delete")}
          </Button>
        </div>
      </section>
      <section className="flex flex-col gap-2 px-4 py-3">
        <h4 className="m-0 text-xs font-semibold text-muted">{t("userGroup.members")}</h4>
        <UserGroupMembers groupId={group.id} workspaceId={workspaceId} />
      </section>
      {dialog === "edit" && (
        <UserGroupDialog
          workspaceId={workspaceId}
          group={group}
          onClose={() => {
            setDialog(null);
          }}
        />
      )}
      <AlertDialog
        isOpen={dialog === "delete"}
        onOpenChange={(isOpen) => {
          setDialog(isOpen ? "delete" : null);
        }}
        title={t("userGroup.delete", { name: group.name })}
        confirmLabel={t("common.delete")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { groupId: group.id },
            {
              onSuccess: () => {
                setDialog(null);
                closePanel();
              },
            },
          );
        }}
      >
        {t("userGroup.deleteBody")}
      </AlertDialog>
    </div>
  );
};
