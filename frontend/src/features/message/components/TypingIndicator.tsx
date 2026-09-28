import { Text } from "@mantine/core";

import { useMembers } from "#/features/member/hooks/useMembers";
import { useWorkspaceId } from "#/lib/routeParams";

type TypingIndicatorProps = {
  userIds: string[];
};

export const TypingIndicator = ({ userIds }: TypingIndicatorProps) => {
  const workspaceId = useWorkspaceId();
  const { data: members } = useMembers(workspaceId);

  if (userIds.length === 0) {
    return null;
  }

  const names = userIds.map(
    (userId) => members?.find((member) => member.userId === userId)?.displayName ?? "誰か",
  );
  const label = names.length > 2 ? `${names.slice(0, 2).join("、")}ほか` : names.join("、");

  return (
    <Text size="xs" c="dimmed" className="px-4 py-1">
      {label}が入力中...
    </Text>
  );
};
