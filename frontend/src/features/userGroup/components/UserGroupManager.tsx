import { useState } from "react";

import { ActionIcon, Button, Card, Group, Stack, Text, TextInput } from "@mantine/core";
import { IconTrash } from "@tabler/icons-react";

import { UserGroupMembers } from "#/features/userGroup/components/UserGroupMembers";
import { useUserGroupActions, useUserGroups } from "#/features/userGroup/hooks/useUserGroups";

type UserGroupManagerProps = {
  workspaceId: string;
};

export const UserGroupManager = ({ workspaceId }: UserGroupManagerProps) => {
  const { data: groups } = useUserGroups(workspaceId);
  const { create, remove } = useUserGroupActions();
  const [name, setName] = useState("");

  return (
    <Stack gap="sm">
      <Text fw={600}>ユーザーグループ ({groups?.length ?? 0})</Text>

      {groups?.map((group) => (
        <Card key={group.id} withBorder padding="sm" radius="md">
          <Group justify="space-between" className="mb-2">
            <div>
              <Text size="sm" fw={600}>
                @{group.name}
              </Text>
              {group.description !== undefined && group.description.length > 0 && (
                <Text size="xs" c="dimmed">
                  {group.description}
                </Text>
              )}
            </div>
            <ActionIcon
              variant="subtle"
              color="red"
              aria-label={`${group.name} を削除`}
              onClick={() => {
                remove.mutate({ groupId: group.id });
              }}
            >
              <IconTrash size={16} />
            </ActionIcon>
          </Group>
          <UserGroupMembers groupId={group.id} workspaceId={workspaceId} />
        </Card>
      ))}

      <Group gap="xs" align="end">
        <TextInput
          size="xs"
          className="flex-1"
          label="グループを作成"
          placeholder="グループ名"
          value={name}
          onChange={(event) => {
            setName(event.currentTarget.value);
          }}
        />
        <Button
          size="xs"
          disabled={name.trim().length === 0}
          loading={create.isPending}
          onClick={() => {
            create.mutate({ name: name.trim(), workspaceId });
            setName("");
          }}
        >
          作成
        </Button>
      </Group>

      {create.isError && (
        <Text c="red" size="xs">
          {create.error.message}
        </Text>
      )}
    </Stack>
  );
};
