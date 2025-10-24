import { useEffect, useState } from "react";

import { Button, Card, Group, Loader, ScrollArea, Stack, Text } from "@mantine/core";
import { useAtomValue, useSetAtom } from "jotai";

import { useChannels } from "../hooks/useChannel";

import { CreateChannelModal } from "./CreateChannelModal";

import { navigateToChannel } from "@/lib/navigation";
import {
  currentChannelIdAtom,
  setCurrentChannelAtom,
} from "@/lib/store/workspace";

type ChannelListProps = {
  workspaceId: string | null;
}

export const ChannelList = ({ workspaceId }: ChannelListProps) => {
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);
  const { data: channels, isLoading, isError, error } = useChannels(workspaceId);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleChannelClick = (channelId: string) => {
    if (workspaceId) {
      navigateToChannel(workspaceId, channelId);
    }
  };

  useEffect(() => {
    if (channels && channels.length > 0 && currentChannelId === null) {
      const firstChannel = channels[0];
      if (firstChannel) {
        setCurrentChannel(firstChannel.id);
      }
    }
  }, [channels, currentChannelId, setCurrentChannel]);

  if (workspaceId === null) {
    return (
      <Card withBorder padding="lg">
        <Text c="dimmed" size="sm">
          ワークスペースを選択するとチャンネルが表示されます
        </Text>
      </Card>
    );
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader size="sm" />
      </div>
    );
  }

  if (isError) {
    return (
      <Card withBorder padding="md">
        <Text c="red" size="sm">
          {error?.message ?? "チャンネルの取得に失敗しました"}
        </Text>
      </Card>
    );
  }

  return (
    <>
      <Stack gap="sm">
        <Group justify="space-between">
          <Text fw={500}>チャンネル</Text>
          <Button size="xs" onClick={() => setIsModalOpen(true)}>
            作成
          </Button>
        </Group>

        {channels && channels.length > 0 ? (
          <ScrollArea h={320} type="auto">
            <Stack gap={4}>
              {channels.map((channel) => {
                const isSelected = channel.id === currentChannelId;
                return (
                  <Button
                    key={channel.id}
                    variant={isSelected ? "filled" : "light"}
                    justify="flex-start"
                    onClick={() => handleChannelClick(channel.id)}
                  >
                    #{channel.name}
                  </Button>
                );
              })}
            </Stack>
          </ScrollArea>
        ) : (
          <Card withBorder padding="md">
            <Text c="dimmed" size="sm">
              チャンネルがありません
            </Text>
            <Button mt="md" size="xs" onClick={() => setIsModalOpen(true)}>
              最初のチャンネルを作成
            </Button>
          </Card>
        )}
      </Stack>

      <CreateChannelModal
        workspaceId={workspaceId}
        opened={isModalOpen}
        onClose={() => setIsModalOpen(false)}
      />
    </>
  );
};
