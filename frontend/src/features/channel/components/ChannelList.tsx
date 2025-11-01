import { useEffect, useState } from "react";

import { Badge, Button, Card, Loader, ScrollArea, Stack, Text } from "@mantine/core";
import { useAtomValue, useSetAtom } from "jotai";

import { useChannels } from "../hooks/useChannel";

import { ChannelName } from "./ChannelName";
import { CreateChannelModal } from "./CreateChannelModal";

import { router } from "@/lib/router";
import { currentChannelIdAtom, setCurrentChannelAtom } from "@/providers/store/workspace";

type ChannelListProps = {
  workspaceId: string | null;
};

export const ChannelList = ({ workspaceId }: ChannelListProps) => {
  const currentChannelId = useAtomValue(currentChannelIdAtom);
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);
  const { data: channels, isLoading, isError, error } = useChannels(workspaceId);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleChannelClick = (channelId: string) => {
    if (workspaceId) {
      setCurrentChannel(channelId);
      router.navigate({ to: "/app/$workspaceId/$channelId", params: { workspaceId, channelId } });
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
      <Stack gap="sm" className="p-2">
        {channels && channels.length > 0 ? (
          <ScrollArea h={320} type="auto">
            <Stack gap={4}>
              {channels.map((channel) => {
                const isSelected = channel.id === currentChannelId;
                const channelData = channel as typeof channel & {
                  hasMention?: boolean;
                  mentionCount?: number;
                };
                const mentionCount = channelData.mentionCount || 0;
                const hasMentionCount = mentionCount > 0;
                // 未読のメッセージがある場合はメンション数が0より大きい場合とする
                const hasUnread = hasMentionCount;

                return (
                  <Button
                    key={channel.id}
                    variant={isSelected ? "filled" : "light"}
                    justify="flex-start"
                    onClick={() => handleChannelClick(channel.id)}
                    className="relative"
                    classNames={{
                      label: "w-full flex items-center justify-between",
                    }}
                  >
                    <ChannelName name={channel.name} isPrivate={channel.isPrivate} isBold={hasUnread} />
                    <div className="flex items-center gap-1">
                      {hasMentionCount ? (
                        <Badge color="blue" size="xs" className="flex items-center justify-center">
                          {mentionCount > 99 ? "99+" : mentionCount}
                        </Badge>
                      ) : null}
                    </div>
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
