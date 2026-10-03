import { useChannelById } from "#/features/channel/hooks/useChannelById";
import { relativePath } from "#/features/channel/utils/channelTree";

import { SystemMessageItem } from "./SystemMessageItem";

import type { SystemMessage } from "#/gen/chat/v1/message_pb";

type AggregatedSystemMessageItemProps = {
  workspaceId: string;
  // 集約表示している親チャンネルのパス
  parentName: string;
  message: SystemMessage;
};

// 下階層もまとめて表示しているときは、どのチャンネルでの出来事かを添える
export const AggregatedSystemMessageItem = ({
  workspaceId,
  parentName,
  message,
}: AggregatedSystemMessageItemProps) => {
  const name = useChannelById(workspaceId, message.channelId).channel?.name;
  return (
    <SystemMessageItem
      message={message}
      channelLabel={name === undefined ? null : relativePath(parentName, name)}
    />
  );
};
