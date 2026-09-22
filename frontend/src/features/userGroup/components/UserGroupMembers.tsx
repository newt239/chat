import { useState } from "react";

import { ActionIcon, Button, Group, Select, Stack, Text } from "@mantine/core";
import { IconUserMinus } from "@tabler/icons-react";

import { useMembers } from "#/features/member/hooks/useMembers";
import {
  useUserGroupMemberActions,
  useUserGroupMembers,
} from "#/features/userGroup/hooks/useUserGroupMembers";

type UserGroupMembersProps = {
  groupId: string;
  workspaceId: string;
};

export const UserGroupMembers = ({ groupId, workspaceId }: UserGroupMembersProps) => {
  const { data: members } = useUserGroupMembers(groupId);
  const { data: workspaceMembers } = useMembers(workspaceId);
  const { add, remove } = useUserGroupMemberActions(groupId);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const memberIds = new Set(members?.map((member) => member.userId));
  const displayNameOf = (userId: string) =>
    workspaceMembers?.find((member) => member.userId === userId)?.displayName ?? userId;

  return (
    <Stack gap="xs">
      {members?.map((member) => (
        <Group key={member.userId} justify="space-between">
          <Text size="sm">{displayNameOf(member.userId)}</Text>
          <ActionIcon
            variant="subtle"
            color="red"
            aria-label={`${displayNameOf(member.userId)} をグループから外す`}
            onClick={() => {
              remove.mutate(member.userId);
            }}
          >
            <IconUserMinus size={16} />
          </ActionIcon>
        </Group>
      ))}

      <Group gap="xs" align="end">
        <Select
          size="xs"
          className="flex-1"
          placeholder="メンバーを追加"
          searchable
          value={selectedUserId}
          onChange={setSelectedUserId}
          data={(workspaceMembers ?? [])
            .filter((member) => !memberIds.has(member.userId))
            .map((member) => ({ label: member.displayName, value: member.userId }))}
        />
        <Button
          size="xs"
          disabled={selectedUserId === null}
          loading={add.isPending}
          onClick={() => {
            if (selectedUserId !== null) {
              add.mutate(selectedUserId);
              setSelectedUserId(null);
            }
          }}
        >
          追加
        </Button>
      </Group>
    </Stack>
  );
};
