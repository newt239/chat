import { Text, Stack, Loader, Badge, Avatar } from "@mantine/core";

import { useMembers } from "@/features/user/hooks/useMembers";

type MemberPanelProps = {
  workspaceId: string | null;
};

export const MemberPanel = ({ workspaceId }: MemberPanelProps) => {
  const { data: members, isLoading, error } = useMembers(workspaceId);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader size="sm" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4">
        <Text c="red" size="sm">
          メンバー情報の読み込みに失敗しました
        </Text>
      </div>
    );
  }

  if (!members || members.length === 0) {
    return (
      <div className="p-4">
        <Text c="dimmed" size="sm">
          メンバーが見つかりませんでした
        </Text>
      </div>
    );
  }

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case "owner":
        return "blue";
      case "admin":
        return "grape";
      case "member":
        return "gray";
      default:
        return "gray";
    }
  };

  const getRoleLabel = (role: string) => {
    switch (role) {
      case "owner":
        return "オーナー";
      case "admin":
        return "管理者";
      case "member":
        return "メンバー";
      default:
        return role;
    }
  };

  return (
    <div className="border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto">
      <Text size="sm" fw={600} className="mb-4">
        メンバー ({members.length})
      </Text>
      <Stack gap="md">
        {members.map((member) => (
          <div key={member.userId} className="flex items-start gap-3">
            <Avatar
              src={member.avatarUrl ?? undefined}
              alt={member.displayName}
              radius="xl"
              size="md"
            >
              {member.displayName.substring(0, 2).toUpperCase()}
            </Avatar>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 mb-1">
                <Text size="sm" fw={500} className="truncate">
                  {member.displayName}
                </Text>
                <Badge size="xs" color={getRoleBadgeColor(member.role)}>
                  {getRoleLabel(member.role)}
                </Badge>
              </div>
              <Text size="xs" c="dimmed" className="truncate">
                {member.email}
              </Text>
            </div>
          </div>
        ))}
      </Stack>
    </div>
  );
};
