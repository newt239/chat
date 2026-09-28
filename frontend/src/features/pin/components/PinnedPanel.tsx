import { Card, ScrollArea, Stack, Text } from "@mantine/core";
import { IconPin } from "@tabler/icons-react";
import { Link } from "@tanstack/react-router";
import { useAtomValue } from "jotai";

import { usePinnedMessages } from "#/features/pin/hooks/usePinnedMessages";
import { toDate } from "#/lib/timestamp";
import { currentWorkspaceIdAtom, currentChannelIdAtom } from "#/providers/store/workspace";

type PinnedPanelProps = {
  channelId?: string | null;
};

export const PinnedPanel = ({ channelId }: PinnedPanelProps) => {
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const effectiveChannelId = channelId ?? currentChannelId;
  const { pins, isLoading, isError, error } = usePinnedMessages(effectiveChannelId);

  if (!workspaceId || !effectiveChannelId) {
    return null;
  }

  if (isLoading) {
    return (
      <div className="p-4">
        <Text size="sm" c="dimmed">
          読み込み中...
        </Text>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="p-4">
        <Text size="sm" c="red">
          {error instanceof Error ? error.message : "ピンの取得に失敗しました"}
        </Text>
      </div>
    );
  }

  if (pins.length === 0) {
    return (
      <div className="p-4 text-center">
        <IconPin size={48} className="mx-auto mb-4 text-gray-400" />
        <Text size="sm" c="dimmed">
          ピン留めされたメッセージはありません
        </Text>
      </div>
    );
  }

  return (
    <ScrollArea h={400}>
      <Stack gap="xs" p="xs">
        {pins.map((pin) => (
          <Card
            key={pin.message.id}
            withBorder
            padding="md"
            radius="md"
            renderRoot={(props) => (
              <Link
                {...props}
                to="/app/$workspaceId/$channelId"
                params={{ channelId: effectiveChannelId, workspaceId }}
                search={{ message: pin.message.id }}
              />
            )}
            className="h-auto text-left justify-start"
          >
            <div className="flex-1 min-w-0">
              <Text size="sm" fw={500} className="whitespace-pre-wrap">
                {pin.message.body}
              </Text>
              <Text size="xs" c="dimmed" mt={4}>
                {toDate(pin.pinnedAt).toLocaleDateString("ja-JP", {
                  day: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                  month: "short",
                  year: "numeric",
                })}
              </Text>
            </div>
          </Card>
        ))}
      </Stack>
    </ScrollArea>
  );
};
