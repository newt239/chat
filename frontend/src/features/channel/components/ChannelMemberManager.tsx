import { useState } from "react";

import { ActionIcon, Avatar, Badge, Button, Group, Select, Stack, Text } from "@mantine/core";
import { IconUserMinus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";

import { useChannelMemberActions } from "#/features/channel/hooks/useChannelMemberActions";
import { useChannelMembers } from "#/features/channel/hooks/useChannelMembers";
import { useMembers } from "#/features/member/hooks/useMembers";
import { userAtom } from "#/providers/store/auth";

import type { components } from "#/lib/api/schema";

type ChannelRole = components["schemas"]["ChannelMemberInfo"]["role"];

const ROLE_OPTIONS: { label: string; value: ChannelRole }[] = [
  { label: "メンバー", value: "member" },
  { label: "管理者", value: "admin" },
];

type ChannelMemberManagerProps = {
  channelId: string;
  workspaceId: string;
};

export const ChannelMemberManager = ({ channelId, workspaceId }: ChannelMemberManagerProps) => {
  const currentUser = useAtomValue(userAtom);
  const { data: channelMembers } = useChannelMembers(channelId);
  const { data: workspaceMembers } = useMembers(workspaceId);
  const { invite, join, leave, remove, updateRole } = useChannelMemberActions(
    channelId,
    workspaceId,
  );

  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);

  const memberIds = new Set(channelMembers?.map((member) => member.userId));
  const isJoined = currentUser !== null && memberIds.has(currentUser.id);

  const invitableMembers = (workspaceMembers ?? []).filter(
    (member) => !memberIds.has(member.userId),
  );

  const handleInvite = () => {
    if (selectedUserId !== null) {
      invite.mutate({ role: "member", userId: selectedUserId });
      setSelectedUserId(null);
    }
  };

  return (
    <Stack gap="sm">
      <Group justify="space-between">
        <Text fw={600}>メンバー ({channelMembers?.length ?? 0})</Text>
        {isJoined ? (
          <Button
            size="xs"
            variant="light"
            color="red"
            loading={leave.isPending}
            onClick={() => {
              leave.mutate();
            }}
          >
            退出する
          </Button>
        ) : (
          <Button
            size="xs"
            variant="light"
            loading={join.isPending}
            onClick={() => {
              join.mutate();
            }}
          >
            参加する
          </Button>
        )}
      </Group>

      <Stack gap="xs">
        {channelMembers?.map((member) => (
          <Group key={member.userId} justify="space-between" wrap="nowrap">
            <Group gap="xs" wrap="nowrap" className="min-w-0">
              <Avatar src={member.avatarUrl ?? undefined} size="sm" radius="xl" />
              <Text size="sm" truncate>
                {member.displayName}
              </Text>
            </Group>
            <Group gap={4} wrap="nowrap">
              <Select
                size="xs"
                w={110}
                data={ROLE_OPTIONS}
                value={member.role}
                allowDeselect={false}
                onChange={(value) => {
                  const role = ROLE_OPTIONS.find((option) => option.value === value);
                  if (role !== undefined) {
                    updateRole.mutate({ role: role.value, userId: member.userId });
                  }
                }}
              />
              <ActionIcon
                variant="subtle"
                color="red"
                aria-label={`${member.displayName} をチャンネルから外す`}
                onClick={() => {
                  remove.mutate({ userId: member.userId });
                }}
              >
                <IconUserMinus size={16} />
              </ActionIcon>
            </Group>
          </Group>
        ))}
        {channelMembers?.length === 0 && (
          <Text size="sm" c="dimmed">
            まだメンバーがいません
          </Text>
        )}
      </Stack>

      <Group gap="xs" align="end">
        <Select
          size="xs"
          className="flex-1"
          label="メンバーを招待"
          placeholder="ユーザーを選択"
          searchable
          value={selectedUserId}
          onChange={setSelectedUserId}
          data={invitableMembers.map((member) => ({
            label: member.displayName,
            value: member.userId,
          }))}
        />
        <Button
          size="xs"
          disabled={selectedUserId === null}
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
      {(remove.isError || updateRole.isError || leave.isError || join.isError) && (
        <Badge color="red" variant="light">
          操作に失敗しました
        </Badge>
      )}
    </Stack>
  );
};
