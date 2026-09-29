import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDMs } from "#/features/dm/hooks/useDM";

/** モバイルのタブとアプリアイコンのバッジに出す未読。DM は未読の合計、通知は未読のメンションがあるチャンネルの数（ミュートを除く） */
export const useUnreadSummary = (workspaceId: string) => {
  const { data: dms = [] } = useDMs(workspaceId);
  const { data: channels = [] } = useChannels(workspaceId);
  return {
    activityUnread: channels.filter((channel) => channel.hasMention && !channel.isMuted).length,
    dmUnread: dms.reduce((sum, dm) => sum + (dm.isMuted ? 0 : dm.unreadCount), 0),
  };
};
