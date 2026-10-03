import { useParams } from "@tanstack/react-router";

import { ChannelHeader } from "#/features/channel/components/ChannelHeader";
import { JoinChannelBar } from "#/features/channel/components/JoinChannelBar";
import { useChannelById } from "#/features/channel/hooks/useChannelById";
import { useViewChannel } from "#/features/channel/hooks/useChannelViewers";
import { useDMs } from "#/features/channel/hooks/useDM";
import { MessageInput } from "#/features/message/components/MessageInput";
import { MessagePanel } from "#/features/message/components/MessagePanel";
import { TypingIndicator } from "#/features/message/components/TypingIndicator";

export const ChannelPage = () => {
  const { workspaceId, channelId } = useParams({ from: "/app/$workspaceId/$channelId" });
  useViewChannel(channelId);
  const { data: dms } = useDMs(workspaceId);
  const dm = dms?.find((candidate) => candidate.id === channelId);
  const channel = useChannelById(workspaceId, dms === undefined || dm ? null : channelId);
  // 一覧にない公開チャンネルや、ツリーをつなぐための未参加の祖先はプレビューとして開く
  const isPreview = channel !== undefined && !channel.isMember;

  return (
    <>
      <ChannelHeader workspaceId={workspaceId} channelId={channelId} channel={channel} dm={dm} />
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
        <div key={channelId} className="relative">
          <TypingIndicator channelId={channelId} />
          <MessageInput channelId={channelId} />
        </div>
      )}
    </>
  );
};
