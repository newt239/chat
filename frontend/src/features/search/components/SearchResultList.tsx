import { Card, Stack, Text, Avatar, Badge } from "@mantine/core";
import { IconHash, IconUser } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";

import type { components } from "@/lib/api/schema";

type Message = components["schemas"]["Message"];
type Channel = components["schemas"]["Channel"];
type MemberInfo = components["schemas"]["MemberInfo"];

interface SearchResultListProps {
  messages: Message[];
  channels: Channel[];
  users: MemberInfo[];
  filter: "all" | "messages" | "channels" | "users";
  workspaceId: string;
}

export const SearchResultList = ({
  messages,
  channels,
  users,
  filter,
  workspaceId,
}: SearchResultListProps) => {
  const navigate = useNavigate();

  const handleChannelClick = (channelId: string) => {
    navigate({
      to: "/app/$workspaceId/$channelId",
      params: { workspaceId, channelId },
    });
  };

  const handleMessageClick = (channelId: string, messageId: string) => {
    navigate({
      to: "/app/$workspaceId/$channelId",
      params: { workspaceId, channelId },
      search: { message: messageId },
    });
  };

  const dateTimeFormatter = new Intl.DateTimeFormat("ja-JP", {
    dateStyle: "short",
    timeStyle: "short",
  });

  return (
    <Stack gap="md">
      {(filter === "all" || filter === "channels") && channels.length > 0 && (
        <div>
          <Text size="sm" fw={600} c="dimmed" className="mb-2">
            チャンネル
          </Text>
          <Stack gap="xs">
            {channels.map((channel) => (
              <Card
                key={channel.id}
                withBorder
                padding="md"
                radius="md"
                className="cursor-pointer hover:bg-gray-50"
                onClick={() => handleChannelClick(channel.id)}
              >
                <div className="flex items-start gap-3">
                  <div className="mt-1">
                    <IconHash size={20} className="text-gray-600" />
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Text size="sm" fw={600}>
                        {channel.name}
                      </Text>
                      {channel.isPrivate && (
                        <Badge size="xs" variant="light" color="gray">
                          プライベート
                        </Badge>
                      )}
                    </div>
                    {channel.description && (
                      <Text size="xs" c="dimmed" className="mt-1">
                        {channel.description}
                      </Text>
                    )}
                  </div>
                </div>
              </Card>
            ))}
          </Stack>
        </div>
      )}

      {(filter === "all" || filter === "users") && users.length > 0 && (
        <div>
          <Text size="sm" fw={600} c="dimmed" className="mb-2">
            ユーザー
          </Text>
          <Stack gap="xs">
            {users.map((user) => (
              <Card key={user.userId} withBorder padding="md" radius="md">
                <div className="flex items-center gap-3">
                  {user.avatarUrl ? (
                    <Avatar src={user.avatarUrl} size="md" radius="xl" />
                  ) : (
                    <Avatar size="md" radius="xl" color="blue">
                      <IconUser size={20} />
                    </Avatar>
                  )}
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Text size="sm" fw={600}>
                        {user.displayName}
                      </Text>
                      <Badge size="xs" variant="light" color="blue">
                        {user.role}
                      </Badge>
                    </div>
                    <Text size="xs" c="dimmed">
                      {user.email}
                    </Text>
                  </div>
                </div>
              </Card>
            ))}
          </Stack>
        </div>
      )}

      {(filter === "all" || filter === "messages") && messages.length > 0 && (
        <div>
          <Text size="sm" fw={600} c="dimmed" className="mb-2">
            メッセージ
          </Text>
          <Stack gap="xs">
            {messages.map((message) => (
              <Card
                key={message.id}
                withBorder
                padding="md"
                radius="md"
                className="cursor-pointer hover:bg-gray-50"
                onClick={() => handleMessageClick(message.channelId, message.id)}
              >
                <div className="flex flex-col gap-2">
                  <div className="flex items-center justify-between">
                    <Text size="xs" c="dimmed">
                      投稿日時: {dateTimeFormatter.format(new Date(message.createdAt))}
                    </Text>
                    {message.editedAt && (
                      <Text size="xs" c="dimmed">
                        (編集済み)
                      </Text>
                    )}
                  </div>
                  <Text size="sm" lineClamp={3}>
                    {message.body}
                  </Text>
                </div>
              </Card>
            ))}
          </Stack>
        </div>
      )}
    </Stack>
  );
};
