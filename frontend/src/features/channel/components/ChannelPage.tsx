import { useParams } from "@tanstack/react-router";

import { ChannelHeader } from "#/features/channel/components/ChannelHeader";
import { JoinChannelBar } from "#/features/channel/components/JoinChannelBar";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { useChannelById } from "#/features/channel/hooks/useChannelById";
import { useViewChannel } from "#/features/channel/hooks/useChannelViewers";
import { useDMs } from "#/features/dm/hooks/useDM";
import { MessageInput } from "#/features/message/components/MessageInput";
import { MessagePanel } from "#/features/message/components/MessagePanel";
import { TypingIndicator } from "#/features/message/components/TypingIndicator";

export const ChannelPage = () => {
  const { workspaceId, channelId } = useParams({ from: "/app/$workspaceId/$channelId" });
  useViewChannel(channelId);
  const { data: channels } = useChannels(workspaceId);
  const { data: dms } = useDMs(workspaceId);
  const isDM = dms?.some((dm) => dm.id === channelId) ?? true;
  const channel = useChannelById(workspaceId, isDM ? null : channelId);
  // 一覧にない公開チャンネルや、ツリーをつなぐための未参加の祖先はプレビューとして開く
  const isPreview =
    channels !== undefined &&
    !isDM &&
    channel !== undefined &&
    !(channels.find((candidate) => candidate.id === channelId)?.isMember ?? false);

  return (
    <>
      <ChannelHeader workspaceId={workspaceId} channelId={channelId} />
      <div className="min-h-0 flex-1">
        <MessagePanel workspaceId={workspaceId} channelId={channelId} />
      </div>
      {isPreview ? (
        <JoinChannelBar
          workspaceId={workspaceId}
          channelId={channelId}
          channelName={channel.name}
        />
      ) : (
        <div className="relative">
          <TypingIndicator key={channelId} channelId={channelId} />
          <MessageInput key={channelId} channelId={channelId} />
        </div>
      )}
    </>
  );
};
