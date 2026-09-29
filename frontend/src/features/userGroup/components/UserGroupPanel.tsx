import { useState } from "react";

import { IconAt, IconEdit, IconTrash } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { LinkButton } from "#/components/ui/LinkButton/LinkButton";
import { toast } from "#/components/ui/ToastRegion/toast";
import { closePanel, openDialog } from "#/features/layout/utils/overlaySearch";

import { useCanManageUserGroups } from "../hooks/useCanManageUserGroups";
import { useUserGroupActions, useUserGroups } from "../hooks/useUserGroups";
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
  const navigate = useNavigate();
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);
  const canManage = useCanManageUserGroups(workspaceId);
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
          {canManage && (
            <LinkButton
              variant="secondary"
              size="sm"
              to="."
              search={openDialog({ dialog: "edit-group" })}
            >
              <IconEdit aria-hidden />
              {t("userGroup.edit")}
            </LinkButton>
          )}
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
          {canManage && (
            <Button
              variant="ghost"
              size="sm"
              className="text-danger"
              onPress={() => {
                setIsDeleteConfirming(true);
              }}
            >
              <IconTrash aria-hidden />
              {t("common.delete")}
            </Button>
          )}
        </div>
      </section>
      <section className="flex flex-col gap-2 px-4 py-3">
        <h4 className="m-0 text-xs font-semibold text-muted">{t("userGroup.members")}</h4>
        <UserGroupMembers groupId={group.id} workspaceId={workspaceId} />
      </section>
      <AlertDialog
        isOpen={isDeleteConfirming}
        onOpenChange={setIsDeleteConfirming}
        title={t("userGroup.delete", { name: group.name })}
        confirmLabel={t("common.delete")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { groupId: group.id },
            {
              onSuccess: () => {
                setIsDeleteConfirming(false);
                void navigate({ search: closePanel, to: "." });
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
