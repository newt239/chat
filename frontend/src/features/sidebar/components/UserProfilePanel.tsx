import { useMemo } from "react";

import { Avatar, Badge, Loader, Stack, Text } from "@mantine/core";

import { useMembers } from "@/features/user/hooks/useMembers";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

type UserProfilePanelProps = {
  workspaceId: string;
  userId: string;
};

export const UserProfilePanel = ({ workspaceId, userId }: UserProfilePanelProps) => {
  const { data: members, isLoading, isError, error } = useMembers(workspaceId);
  const member = useMemo(() => {
    if (members === undefined) {
      return null;
    }
    return members.find((candidate) => candidate.userId === userId) ?? null;
  }, [members, userId]);

  if (isLoading) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (isError) {
    const message =
      error instanceof Error ? error.message : "ユーザープロフィールの取得に失敗しました";
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (member === null) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          指定されたユーザーが見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div className="flex items-center gap-3">
          <Avatar src={member.avatarUrl ?? undefined} radius="xl" size="lg">
            {member.displayName.substring(0, 2).toUpperCase()}
          </Avatar>
          <div>
            <Text size="sm" fw={600}>
              {member.displayName}
            </Text>
            <Text size="xs" c="dimmed">
              {member.email}
            </Text>
          </div>
        </div>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ロール
          </Text>
          <Badge size="sm" variant="light" color="gray">
            {member.role}
          </Badge>
        </Stack>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ユーザーID
          </Text>
          <Text size="xs" c="dimmed">
            {member.userId}
          </Text>
        </Stack>
      </Stack>
    </div>
  );
};
