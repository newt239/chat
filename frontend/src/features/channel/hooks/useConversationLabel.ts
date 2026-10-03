import { useDMs } from "#/features/channel/hooks/useDM";
import { dmName } from "#/features/channel/utils/dmName";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";

import { useChannels } from "./useChannel";

// 一覧の見出しに出す会話の名前。チャンネルは #名前、DM は相手の名前。分からなければ空文字
export const useConversationLabel = (workspaceId: string) => {
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const displayName = useDisplayName();
  return (channelId: string) => {
    const channel = channels?.find((item) => item.id === channelId);
    if (channel) {
      return `#${channel.name}`;
    }
    const dm = dms?.find((item) => item.id === channelId);
    return dm ? dmName(dm, displayName) : "";
  };
};
