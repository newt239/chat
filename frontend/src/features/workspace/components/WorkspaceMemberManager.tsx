import { useState } from "react";

import { ActionIcon, Avatar, Button, Group, Select, Stack, Text, TextInput } from "@mantine/core";
import { IconUserMinus } from "@tabler/icons-react";

import { useMembers } from "#/features/member/hooks/useMembers";
import { useWorkspaceMemberActions } from "#/features/workspace/hooks/useWorkspaceMemberActions";
import { workspaceRoleLabels } from "#/features/workspace/utils/workspaceRole";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

const ROLE_OPTIONS = [WorkspaceRole.MEMBER, WorkspaceRole.ADMIN].map((role) => ({
  label: workspaceRoleLabels[role],
  role,
  value: String(role),
}));

type WorkspaceMemberManagerProps = {
  workspaceId: string;
};

export const WorkspaceMemberManager = ({ workspaceId }: WorkspaceMemberManagerProps) => {
  const { data: members } = useMembers(workspaceId);
  const { invite, remove, updateRole } = useWorkspaceMemberActions();
  const [email, setEmail] = useState("");

  const handleInvite = () => {
    if (email.trim().length > 0) {
      invite.mutate({ email: email.trim(), role: WorkspaceRole.MEMBER, workspaceId });
      setEmail("");
    }
  };

  return (
    <Stack gap="sm">
      <Text fw={600}>メンバー ({members?.length ?? 0})</Text>

      <Stack gap="xs">
        {members?.map((member) => (
          <Group key={member.userId} justify="space-between" wrap="nowrap">
            <Group gap="xs" wrap="nowrap" className="min-w-0">
              <Avatar src={member.avatarUrl ?? undefined} size="sm" radius="xl" />
              <div className="min-w-0">
                <Text size="sm" truncate>
                  {member.displayName}
                </Text>
                <Text size="xs" c="dimmed" truncate>
                  {member.email}
                </Text>
              </div>
            </Group>
            <Group gap={4} wrap="nowrap">
              {member.role === WorkspaceRole.OWNER ? (
                <Text size="xs" c="dimmed">
                  {workspaceRoleLabels[member.role]}
                </Text>
              ) : (
                <>
                  <Select
                    size="xs"
                    w={110}
                    data={ROLE_OPTIONS}
                    value={String(member.role)}
                    allowDeselect={false}
                    onChange={(value) => {
                      const role = ROLE_OPTIONS.find((option) => option.value === value);
                      if (role !== undefined) {
                        updateRole.mutate({ role: role.role, userId: member.userId, workspaceId });
                      }
                    }}
                  />
                  <ActionIcon
                    variant="subtle"
                    color="red"
                    aria-label={`${member.displayName} をワークスペースから外す`}
                    onClick={() => {
                      remove.mutate({ userId: member.userId, workspaceId });
                    }}
                  >
                    <IconUserMinus size={16} />
                  </ActionIcon>
                </>
              )}
            </Group>
          </Group>
        ))}
      </Stack>

      <Group gap="xs" align="end">
        <TextInput
          size="xs"
          className="flex-1"
          label="メールアドレスで招待"
          placeholder="email@example.com"
          value={email}
          onChange={(event) => {
            setEmail(event.currentTarget.value);
          }}
        />
        <Button
          size="xs"
          disabled={email.trim().length === 0}
          loading={invite.isPending}
          onClick={handleInvite}
        >
          招待
        </Button>
      </Group>

      {invite.isError && (
        <Text c="red" size="xs">
          {invite.error.message}
        </Text>
      )}
    </Stack>
  );
};
