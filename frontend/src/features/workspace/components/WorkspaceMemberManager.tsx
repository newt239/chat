import { IconUserMinus } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { IconButton } from "#/components/ui/IconButton";
import { Select } from "#/components/ui/Select";
import { useMembers } from "#/features/member/hooks/useMembers";
import { workspaceRoleKeys } from "#/features/member/utils/workspaceRoleKeys";
import { useWorkspaceMemberActions } from "#/features/workspace/hooks/useWorkspaceMemberActions";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

import { InviteMemberForm } from "./InviteMemberForm";

const ROLES = [WorkspaceRole.MEMBER, WorkspaceRole.ADMIN];

type WorkspaceMemberManagerProps = {
  workspaceId: string;
};

export const WorkspaceMemberManager = ({ workspaceId }: WorkspaceMemberManagerProps) => {
  const { t } = useTranslation();
  const { data: members = [] } = useMembers(workspaceId);
  const { remove, updateRole } = useWorkspaceMemberActions();
  const roleOptions = ROLES.map((role) => ({
    label: t(workspaceRoleKeys[role]),
    value: String(role),
  }));

  return (
    <section className="flex flex-col gap-2">
      <h3 className="m-0 text-body-strong">
        {t("workspace.members.title", { count: members.length })}
      </h3>
      <ul className="m-0 flex list-none flex-col gap-1.5 p-0">
        {members.map((member) => (
          <li key={member.userId} className="flex items-center gap-2">
            <Avatar name={member.displayName} src={member.avatarUrl} size={28} />
            <span className="flex min-w-0 flex-1 flex-col">
              <span className="truncate text-body">{member.displayName}</span>
              <span className="truncate text-caption text-muted">{member.email}</span>
            </span>
            {member.role === WorkspaceRole.OWNER ? (
              <span className="text-caption text-muted">{t(workspaceRoleKeys[member.role])}</span>
            ) : (
              <>
                <Select
                  label={t("workspace.members.role")}
                  className="w-28 [&>label]:sr-only"
                  options={roleOptions}
                  value={String(member.role)}
                  onChange={(value) => {
                    const role = ROLES.find((candidate) => String(candidate) === value);
                    if (role !== undefined) {
                      updateRole.mutate({ role, userId: member.userId, workspaceId });
                    }
                  }}
                />
                <IconButton
                  label={t("workspace.members.remove", { name: member.displayName })}
                  onPress={() => {
                    remove.mutate({ userId: member.userId, workspaceId });
                  }}
                >
                  <IconUserMinus />
                </IconButton>
              </>
            )}
          </li>
        ))}
      </ul>
      <InviteMemberForm workspaceId={workspaceId} />
    </section>
  );
};
