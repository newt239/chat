import { useMutation } from "@connectrpc/connect-query";
import { Card, Group, Stack, Text, Badge } from "@mantine/core";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ThreadService } from "#/gen/chat/v1/thread_service_pb";
import { toDate } from "#/lib/timestamp";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread } from "#/gen/chat/v1/thread_service_pb";

type ThreadCardProps = {
  thread: ParticipatingThread & { firstMessage: Message };
  onMarkedRead: (threadId: string) => void;
};

export const ThreadCard = ({ thread, onMarkedRead }: ThreadCardProps) => {
  const navigate = useNavigate();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);
  const markThreadRead = useMutation(ThreadService.method.markThreadRead);

  const handleOpenThread = async () => {
    await markThreadRead.mutateAsync({ threadId: thread.threadId });
    onMarkedRead(thread.threadId);

    if (thread.channelId !== undefined) {
      await navigate({
        params: { channelId: thread.channelId, workspaceId },
        search: { message: thread.firstMessage.id },
        to: "/app/$workspaceId/$channelId",
      });
    }
    setRightSidePanelView({ threadId: thread.threadId, type: "thread" });
  };

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
            最終更新: {toDate(thread.lastActivityAt).toLocaleString()}
          </Text>
          <Group gap={8}>
            {thread.unreadCount > 0 && <Badge color="blue">未読 {thread.unreadCount}</Badge>}
            <Badge variant="light">返信 {thread.replyCount}</Badge>
          </Group>
        </Group>
        <Text fw={600}>{thread.firstMessage.body}</Text>
        <Text size="xs" c="dimmed">
          投稿: {toDate(thread.firstMessage.createdAt).toLocaleString()}
        </Text>
      </Stack>
    </Card>
  );
};
