import { useMemo } from "react";

import { Badge, Loader, Stack, Text } from "@mantine/core";
import { useAtomValue } from "jotai";

import { useChannels } from "@/features/channel/hooks/useChannel";
import { currentChannelIdAtom } from "@/lib/store/workspace";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

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
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <div className="flex h-full items-center justify-center">
          <Loader size="sm" />
        </div>
      </div>
    );
  }

  if (isError) {
    const message = error instanceof Error ? error.message : "チャンネル情報の取得に失敗しました";
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text c="red" size="sm">
          {message}
        </Text>
      </div>
    );
  }

  if (!activeChannel) {
    return (
      <div className={SIDEBAR_CONTAINER_CLASS}>
        <Text size="sm" c="dimmed">
          チャンネル情報が見つかりませんでした
        </Text>
      </div>
    );
  }

  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div>
          <Text size="sm" fw={600} className="mb-1">
            チャンネル情報
          </Text>
          <Text size="xs" c="dimmed">
            #{activeChannel.name}
          </Text>
        </div>
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
