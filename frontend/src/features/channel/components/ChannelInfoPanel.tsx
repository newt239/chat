import { useMemo } from "react";

import { Badge, Loader, Stack, Text } from "@mantine/core";
import { useAtomValue } from "jotai";

import { useChannels } from "@/features/channel/hooks/useChannel";
import { currentChannelIdAtom } from "@/providers/store/workspace";

type ChannelInfoPanelProps = {
  workspaceId: string;
  channelId?: string | null;
};

export const ChannelInfoPanel = ({ workspaceId, channelId }: ChannelInfoPanelProps) => {
  const { data: channels, isLoading, isError, error } = useChannels(workspaceId);
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const effectiveChannelId = channelId ?? currentChannelId;
  const activeChannel = useMemo(() => {
    if (channels === undefined || effectiveChannelId === null) {
      return null;
    }
    return channels.find((candidate) => candidate.id === effectiveChannelId) ?? null;
  }, [channels, effectiveChannelId]);

  if (isLoading) {
    return (
      <div>
        <Loader size="sm" />
      </div>
    );
  }

  if (isError) {
    const message = error instanceof Error ? error.message : "チャンネル情報の取得に失敗しました";
    return (
      <div>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (!activeChannel) {
    return (
      <div>
        <Text size="sm" c="dimmed">
          チャンネル情報が見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className="p-4">
      <Stack gap="md">
        <Text size="sm" fw={600} className="mb-1">
          #{activeChannel.name}
        </Text>
        {typeof activeChannel.description === "string" && activeChannel.description.length > 0 ? (
          <Text size="sm">{activeChannel.description}</Text>
        ) : (
          <Text size="sm" c="dimmed">
            説明は設定されていません
          </Text>
        )}
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            ステータス
          </Text>
          <Badge size="sm" variant="light" color={activeChannel.isPrivate ? "gray" : "blue"}>
            {activeChannel.isPrivate ? "プライベート" : "パブリック"}
          </Badge>
        </Stack>
        <Stack gap="xs">
          <Text size="sm" fw={600}>
            チャンネルID
          </Text>
          <Text size="xs" c="dimmed">
            {activeChannel.id}
          </Text>
        </Stack>
      </Stack>
    </div>
  );
};
