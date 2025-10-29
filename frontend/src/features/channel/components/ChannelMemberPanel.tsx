import { Text, Stack, Loader, Avatar, Box } from "@mantine/core";
import { useSetAtom } from "jotai";

import { useChannelMembers } from "@/features/channel/hooks/useChannelMembers";
import { setRightSidePanelViewAtom } from "@/providers/store/ui";

type ChannelMemberPanelProps = {
  channelId: string;
};

export const ChannelMemberPanel = ({ channelId }: ChannelMemberPanelProps) => {
  const { data: members, isLoading, error } = useChannelMembers(channelId);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  const handleUserClick = (userId: string) => {
    setRightSidePanelView({ type: "user-profile", userId });
  };

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

  return (
    <div className="h-full overflow-y-auto">
      <Stack gap={0}>
        {members.map((member) => (
          <Box
            key={member.userId}
            className="flex items-start gap-3 cursor-pointer hover:bg-gray-100 p-2 transition-colors"
            onClick={() => handleUserClick(member.userId)}
          >
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
              </div>
              <Text size="xs" c="dimmed" className="truncate">
                {member.email}
              </Text>
            </div>
          </Box>
        ))}
      </Stack>
    </div>
  );
};
