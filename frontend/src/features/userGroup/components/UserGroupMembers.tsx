import { useState } from "react";

import { IconUserMinus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { MemberPickerForm } from "#/features/member/components/MemberPickerForm";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMembers } from "#/features/member/hooks/useMembers";
import {
  useUserGroupMemberActions,
  useUserGroupMembers,
} from "#/features/userGroup/hooks/useUserGroupMembers";
import { useIsWorkspaceAdmin } from "#/hooks/useIsWorkspaceAdmin";
import { myUserIdAtom } from "#/providers/store/auth";

type UserGroupMembersProps = {
  groupId: string;
  workspaceId: string;
};

export const UserGroupMembers = ({ groupId, workspaceId }: UserGroupMembersProps) => {
  const { t } = useTranslation();
  const { data: members } = useUserGroupMembers(groupId);
  const { data: workspaceMembers } = useMembers(workspaceId);
  const { add, remove } = useUserGroupMemberActions();
  const canManage = useIsWorkspaceAdmin(workspaceId);
  const displayName = useDisplayName();
  const myId = useAtomValue(myUserIdAtom);
  const [removing, setRemoving] = useState<{ userId: string; name: string } | null>(null);

  const memberIds = new Set(members?.map((member) => member.userId));

  return (
    <div className="flex flex-col gap-2">
      {members?.length === 0 && (
        <p className="m-0 text-caption text-muted">{t("userGroup.noMembers")}</p>
      )}
      <ul className="m-0 flex list-none flex-col p-0">
        {members?.map(({ userId }) => {
          const name = displayName(userId, userId);
          const avatarUrl = workspaceMembers?.find(
            (candidate) => candidate.userId === userId,
          )?.avatarUrl;
          return (
            <li key={userId} className="flex items-center gap-2.5 py-0.5 text-body-sm">
              <Avatar name={name} src={avatarUrl} size={24} />
              <span className="min-w-0 flex-1 truncate">{name}</span>
              {/* 管理者でなくても自分はグループから抜けられる */}
              {(canManage || userId === myId) && (
                <IconButton
                  label={t("userGroup.removeMember", { name })}
                  onPress={() => {
                    setRemoving({ name, userId });
                  }}
                >
                  <IconUserMinus />
                </IconButton>
              )}
            </li>
          );
        })}
      </ul>
      {canManage && (
        <MemberPickerForm
          workspaceId={workspaceId}
          memberIds={memberIds}
          label={t("userGroup.addMember")}
          placeholder={t("userGroup.addMemberPlaceholder")}
          submitLabel={t("userGroup.add")}
          isPending={add.isPending}
          onSubmit={(userId) => {
            add.mutate({ groupId, userId });
          }}
        />
      )}
      <AlertDialog
        isOpen={removing !== null}
        onOpenChange={() => {
          setRemoving(null);
        }}
        title={t("userGroup.removeMemberTitle", { name: removing?.name ?? "" })}
        confirmLabel={t("userGroup.remove")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          if (removing === null) {
            return;
          }
          remove.mutate(
            { groupId, userId: removing.userId },
            {
              onSettled: () => {
                setRemoving(null);
              },
            },
          );
        }}
      >
        <p className="m-0">{t("userGroup.removeMemberBody")}</p>
      </AlertDialog>
    </div>
  );
};
