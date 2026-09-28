import { Card, Group, Stack, Text, Badge } from "@mantine/core";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { api } from "#/lib/api/client";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import type { ParticipatingThread } from "#/features/thread/schemas";

type ThreadCardProps = {
  thread: ParticipatingThread;
  onMarkedRead?: (threadId: string) => void;
};

export const ThreadCard = ({ thread, onMarkedRead }: ThreadCardProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  const handleOpenThread = async () => {
    await api.POST("/api/threads/{threadId}/read", {
      params: { path: { threadId: thread.thread_id } },
    });
    onMarkedRead?.(thread.thread_id);

    if (thread.channel_id) {
      await navigate({
        params: { channelId: thread.channel_id, workspaceId },
        search: { message: thread.first_message.id },
        to: "/app/$workspaceId/$channelId",
      });
    }
    setRightSidePanelView({ threadId: thread.thread_id, type: "thread" });
  };

  const first = thread.first_message;

  return (
    <Card
      withBorder
      padding="sm"
      className="cursor-pointer hover:bg-gray-50"
      onClick={() => {
        void handleOpenThread();
      }}
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
