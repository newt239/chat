import { useState } from "react";

import { IconUserMinus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { ComboBox } from "#/components/ui/ComboBox/ComboBox";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { useMembers } from "#/features/member/hooks/useMembers";
import {
  useUserGroupMemberActions,
  useUserGroupMembers,
} from "#/features/userGroup/hooks/useUserGroupMembers";
import { useMyWorkspaceRole } from "#/hooks/useMyWorkspaceRole";
import { isAdminRole } from "#/lib/isAdminRole";
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
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const canManage = isAdminRole(useMyWorkspaceRole(workspaceId).data);
  const myId = useAtomValue(myUserIdAtom);

  const memberIds = new Set(members?.map((member) => member.userId));
  const options = (workspaceMembers ?? [])
    .filter((member) => !memberIds.has(member.userId))
    .map((member) => ({ label: member.nickname ?? member.displayName, value: member.userId }));

  return (
    <div className="flex flex-col gap-2">
      {members?.length === 0 && (
        <p className="m-0 text-caption text-muted">{t("userGroup.noMembers")}</p>
      )}
      <ul className="m-0 flex list-none flex-col p-0">
        {members?.map(({ userId }) => {
          const member = workspaceMembers?.find((candidate) => candidate.userId === userId);
          const name = member?.nickname ?? member?.displayName ?? userId;
          return (
            <li key={userId} className="flex items-center gap-2.5 py-0.5 text-[13.5px]">
              <Avatar name={name} src={member?.avatarUrl} size={24} />
              <span className="min-w-0 flex-1 truncate">{name}</span>
              {/* 管理者でなくても自分はグループから抜けられる */}
              {(canManage || userId === myId) && (
                <IconButton
                  label={t("userGroup.removeMember", { name })}
                  onPress={() => {
                    remove.mutate({ groupId, userId });
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
        <Form
          className="flex items-end gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            if (selectedUserId !== null) {
              add.mutate({ groupId, userId: selectedUserId });
              setSelectedUserId(null);
            }
          }}
        >
          <ComboBox
            label={t("userGroup.addMember")}
            placeholder={t("userGroup.addMemberPlaceholder")}
            options={options}
            value={selectedUserId}
            onChange={setSelectedUserId}
            className="flex-1"
          />
          <Button
            type="submit"
            variant="secondary"
            isDisabled={selectedUserId === null}
            isPending={add.isPending}
          >
            {t("userGroup.add")}
          </Button>
        </Form>
      )}
    </div>
  );
};
