import { useParams } from "@tanstack/react-router";

import { useChannels } from "./useChannel";
import { useChannelById } from "./useChannelById";
import { useDMs } from "./useDM";

// 開いているチャンネルが一覧になく未参加なら、プレビュー中のチャンネルとして返す
export const usePreviewChannel = (workspaceId: string) => {
  const { channelId } = useParams({ strict: false });
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const isListed =
    channels?.some((channel) => channel.id === channelId) === true ||
    dms?.some((dm) => dm.id === channelId) === true;
  const { channel } = useChannelById(
    workspaceId,
    channelId === undefined || dms === undefined || isListed ? null : channelId,
  );
  return channel !== undefined && !channel.isMember ? channel : undefined;
};
