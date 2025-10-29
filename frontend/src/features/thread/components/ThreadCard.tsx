import { Card, Group, Stack, Text, Badge } from "@mantine/core";
import { useNavigate, useParams } from "@tanstack/react-router";

import type { ParticipatingThread } from "@/features/thread/schemas";

import { api } from "@/lib/api/client";

type ThreadCardProps = {
  thread: ParticipatingThread;
  onMarkedRead?: (threadId: string) => void;
};

export const ThreadCard = ({ thread, onMarkedRead }: ThreadCardProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useParams({ from: "/app/$workspaceId/threads" });

  const handleOpenThread = async () => {
    // 既読更新
    await api.POST("/api/threads/{threadId}/read", {
      params: { path: { threadId: thread.thread_id } },
    });
    onMarkedRead?.(thread.thread_id);

    // チャンネルへ遷移（スレッド起点メッセージへは既存の右パネルThreadを使う想定ならここで開いてもよいが、まずはチャンネルへ）
    if (thread.channel_id) {
      navigate({
        to: "/app/$workspaceId/$channelId",
        params: { workspaceId, channelId: thread.channel_id },
      });
    }
  };

  const first = thread.first_message;

  return (
    <Card
      withBorder
      padding="sm"
      className="cursor-pointer hover:bg-gray-50"
      onClick={handleOpenThread}
    >
      <Stack gap={6}>
        <Group justify="space-between" align="center">
          <Text size="sm" c="dimmed">
            最終更新: {new Date(thread.last_activity_at).toLocaleString()}
          </Text>
          <Group gap={8}>
            {thread.unread_count > 0 && <Badge color="blue">未読 {thread.unread_count}</Badge>}
            <Badge variant="light">返信 {thread.reply_count}</Badge>
          </Group>
        </Group>
        <Text fw={600}>{first.body}</Text>
        <Text size="xs" c="dimmed">
          投稿: {new Date(first.createdAt).toLocaleString()}
        </Text>
      </Stack>
    </Card>
  );
};
