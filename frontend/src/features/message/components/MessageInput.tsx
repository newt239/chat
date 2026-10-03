import { useState } from "react";

import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PostTargetPicker } from "#/features/channel/components/PostTargetPicker";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";

import { BaseMessageInput } from "./BaseMessageInput";

type MessageInputProps = {
  channelId: string;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const { t } = useTranslation();
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

  return (
    <BaseMessageInput
      placeholder={
        target
          ? t("channel.aggregate.placeholder", { name: target.name })
          : t("message.composer.placeholder")
      }
      channelId={target?.id ?? channelId}
      parentId={null}
      onSent={null}
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
