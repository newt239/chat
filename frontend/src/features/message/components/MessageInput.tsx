import { useCallback, useState } from "react";

import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { PostTargetPicker } from "#/features/channel/components/PostTargetPicker";
import { useChannelAggregation } from "#/features/channel/hooks/useChannelAggregation";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import { useSendMessage } from "../hooks/useMessage";
import { BaseMessageInput } from "./BaseMessageInput";

import type { ComposerContent } from "../utils/composerContent";

type MessageInputProps = {
  channelId: string | null;
};

export const MessageInput = ({ channelId }: MessageInputProps) => {
  const { t } = useTranslation();
  const sendMessage = useSendMessage();
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
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

  const handleSubmit = useCallback(
    (content: ComposerContent) => {
      if (targetId !== null) {
        sendMessage.mutate({ ...content, channelId: targetId });
      }
    },
    [sendMessage, targetId],
  );

  if (!channelId || targetId === null) {
    return null;
  }

  return (
    <BaseMessageInput
      key={channelId}
      onSubmit={handleSubmit}
      placeholder={
        target
          ? t("channel.aggregate.placeholder", { name: target.name })
          : t("message.composer.placeholder")
      }
      isPending={sendMessage.isPending}
      error={sendMessage.isError ? sendMessage.error.message : undefined}
      channelId={targetId}
      targetPicker={
        target &&
        channel && (
          <PostTargetPicker
            parent={channel}
            descendants={descendants}
            value={target.id}
            onChange={setSelectedId}
          />
        )
      }
    />
  );
};
