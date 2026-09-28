import { useMemo } from "react";

import { Avatar, Badge, Button, Loader, Stack, Text } from "@mantine/core";
import { IconMessage } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { useCreateDM } from "#/features/dm/hooks/useDM";
import { useMembers } from "#/features/member/hooks/useMembers";
import { userAtom } from "#/providers/store/auth";

type UserProfilePanelProps = {
  workspaceId: string;
  userId: string;
};

export const UserProfilePanel = ({ workspaceId, userId }: UserProfilePanelProps) => {
  const { data: members, isLoading, isError, error } = useMembers(workspaceId);
  const currentUser = useAtomValue(userAtom);
  const createDM = useCreateDM(workspaceId);
  const navigate = useNavigate();

  const handleStartDM = async () => {
    const dm = await createDM.mutateAsync({ userId });
    void navigate({
      params: { channelId: dm.id, workspaceId },
      to: "/app/$workspaceId/$channelId",
    });
  };
  const member = useMemo(() => {
    if (members === undefined) {
      return null;
    }
    return members.find((candidate) => candidate.userId === userId) ?? null;
  }, [members, userId]);

  if (isLoading) {
    return (
      <div>
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
      <div>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (member === null) {
    return (
      <div>
        <Text size="sm" c="dimmed">
          指定されたユーザーが見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className="p-4">
      <Stack gap="md">
        <div className="flex items-center gap-3">
          <Avatar src={member.avatarUrl ?? undefined} radius="xl" size="lg">
            {member.displayName.slice(0, 2).toUpperCase()}
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
        {currentUser?.id !== member.userId && (
          <Button
            leftSection={<IconMessage size={16} />}
            variant="light"
            loading={createDM.isPending}
            onClick={() => {
              void handleStartDM();
            }}
          >
            DM を開始
          </Button>
        )}
        {typeof member.bio === "string" && member.bio.length > 0 && (
          <Stack gap="xs">
            <Text size="sm" fw={600}>
              自己紹介
            </Text>
            <Text size="sm">{member.bio}</Text>
          </Stack>
        )}
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
