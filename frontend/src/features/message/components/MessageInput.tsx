import { useState } from "react";

import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PostTargetPicker } from "#/features/channel/components/PostTargetPicker";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";

import { useSendMessage } from "../hooks/useMessage";
import { BaseMessageInput } from "./BaseMessageInput";

import type { ComposerContent } from "./BaseMessageInput";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const { t } = useTranslation();
  const sendMessage = useSendMessage();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { channel, descendants, includesDescendants } = useChannelAggregation(
    workspaceId,
    channelId,
  );
  const [selectedId, setSelectedId] = useState(channelId);
  // 集約表示をやめたり子孫がなくなったりしたら親チャンネルに戻す
  const target =
    includesDescendants && channel
      ? ([channel, ...descendants].find((candidate) => candidate.id === selectedId) ?? channel)
      : null;
  const targetId = target?.id ?? channelId;

  const handleSubmit = (content: ComposerContent) => {
    if (targetId !== null) {
      sendMessage.mutate({ ...content, channelId: targetId });
    }
  };

  if (!channelId || targetId === null) {
    return null;
  }

  return (
    <BaseMessageInput
      onSubmit={handleSubmit}
      placeholder={
        target
          ? t("channel.aggregate.placeholder", { name: target.name })
          : t("message.composer.placeholder")
      }
      isPending={sendMessage.isPending}
      error={sendMessage.isError ? sendMessage.error.message : null}
      channelId={targetId}
      parentId={null}
      targetPicker={
        target &&
        channel && (
          <PostTargetPicker
            parent={channel}
            descendants={descendants}
            value={target}
            onChange={setSelectedId}
          />
        )
      }
    />
  );
};
