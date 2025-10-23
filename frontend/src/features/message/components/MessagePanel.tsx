import { useMemo, useState, useEffect, useRef } from "react";
import type { FormEvent } from "react";

import { Button, Card, Loader, Stack, Text, Textarea } from "@mantine/core";

import { useMessages, useSendMessage } from "../hooks/useMessage";

import { useChannels } from "@/features/channel/hooks/useChannel";

interface MessagePanelProps {
  workspaceId: string | null;
  channelId: string | null;
}

export const MessagePanel = ({ workspaceId, channelId }: MessagePanelProps) => {
  const { data: channels } = useChannels(workspaceId);
  const { data: messageResponse, isLoading, isError, error } = useMessages(channelId);
  const sendMessage = useSendMessage(channelId);
  const [body, setBody] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const dateTimeFormatter = useMemo(
    () =>
      new Intl.DateTimeFormat("ja-JP", {
        dateStyle: "short",
        timeStyle: "short",
      }),
    []
  );
  const orderedMessages = useMemo(() => {
    if (!messageResponse) {
      return [];
    }
    return [...messageResponse.messages].sort((first, second) => {
      const firstTime = new Date(first.createdAt).getTime();
      const secondTime = new Date(second.createdAt).getTime();
      return firstTime - secondTime;
    });
  }, [messageResponse]);

  const activeChannel = useMemo(() => {
    if (!channels || channelId === null) {
      return null;
    }
    return channels.find((channel) => channel.id === channelId) ?? null;
  }, [channels, channelId]);

  // メッセージが読み込まれた時と新しいメッセージが送信された時に最新メッセージにスクロール
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    if (messageResponse && !isLoading) {
      scrollToBottom();
    }
  }, [messageResponse, isLoading]);

  useEffect(() => {
    if (sendMessage.isSuccess) {
      scrollToBottom();
    }
  }, [sendMessage.isSuccess]);

  if (workspaceId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">ワークスペースを選択してください</Text>
      </Card>
    );
  }

  if (channelId === null) {
    return (
      <Card withBorder padding="xl" radius="md" className="h-full flex items-center justify-center">
        <Text c="dimmed">チャンネルを選択するとメッセージが表示されます</Text>
      </Card>
    );
  }

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (body.trim().length === 0) {
      return;
    }
    sendMessage.mutate(
      { body: body.trim() },
      {
        onSuccess: () => {
          setBody("");
        },
      }
    );
  };

  return (
    <div className="flex h-full min-h-0 flex-col">
      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <Stack gap="xs">
          <Text fw={600} size="lg">
            {activeChannel ? `#${activeChannel.name}` : "チャンネル"}
          </Text>
          {activeChannel?.description && (
            <Text size="sm" c="dimmed">
              {activeChannel.description}
            </Text>
          )}
        </Stack>
      </Card>

      <div className="flex-1 overflow-y-auto min-h-0">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <Loader size="sm" />
          </div>
        ) : isError ? (
          <Text c="red" size="sm">
            {error?.message ?? "メッセージの取得に失敗しました"}
          </Text>
        ) : messageResponse && messageResponse.messages.length > 0 ? (
          <div className="flex h-full flex-col">
            {messageResponse.hasMore && (
              <Text size="xs" c="dimmed" className="px-4 py-2 text-center">
                さらに過去のメッセージがあります
              </Text>
            )}
            <div className="flex flex-1 flex-col justify-end">
              <div className="space-y-3 px-4 pb-6">
                {orderedMessages.map((message) => (
                  <div
                    key={message.id}
                    className="group rounded-md px-4 py-2 transition-colors hover:bg-gray-50"
                  >
                    <div className="flex flex-wrap items-baseline gap-2">
                      <Text size="xs" c="dimmed">
                        {dateTimeFormatter.format(new Date(message.createdAt))}
                      </Text>
                    </div>
                    <Text className="mt-1 whitespace-pre-wrap break-words text-sm leading-relaxed text-gray-900">
                      {message.body}
                    </Text>
                  </div>
                ))}
                <div ref={messagesEndRef} />
              </div>
            </div>
          </div>
        ) : (
          <div className="flex h-full items-center justify-center">
            <Text c="dimmed" size="sm">
              メッセージはまだありません
            </Text>
          </div>
        )}
      </div>

      <Card withBorder padding="lg" radius="md" className="shrink-0">
        <form onSubmit={handleSubmit}>
          <Textarea
            placeholder="メッセージを入力..."
            minRows={3}
            autosize
            value={body}
            onChange={(event) => setBody(event.currentTarget.value)}
            disabled={sendMessage.isPending}
          />
          {sendMessage.isError && (
            <Text c="red" size="sm" className="mt-2">
              {sendMessage.error?.message ?? "メッセージの送信に失敗しました"}
            </Text>
          )}
          <Button
            type="submit"
            className="mt-2"
            disabled={body.trim().length === 0}
            loading={sendMessage.isPending}
          >
            送信
          </Button>
        </form>
      </Card>
    </div>
  );
};
