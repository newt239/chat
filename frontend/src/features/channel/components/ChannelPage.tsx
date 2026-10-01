import { useEffect } from "react";

import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ChannelHeader } from "#/features/channel/components/ChannelHeader";
import { JoinChannelBar } from "#/features/channel/components/JoinChannelBar";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { useChannelById } from "#/features/channel/hooks/useChannelById";
import { useViewChannel } from "#/features/channel/hooks/useChannelViewers";
import { useDMs } from "#/features/dm/hooks/useDM";
import { MessageInput } from "#/features/message/components/MessageInput";
import { MessagePanel } from "#/features/message/components/MessagePanel";
import { TypingIndicator } from "#/features/message/components/TypingIndicator";
import { MiniPlayer } from "#/features/player/components/MiniPlayer";
import { setCurrentChannelAtom } from "#/providers/store/workspace";

export const ChannelPage = () => {
  const { workspaceId, channelId } = useParams({ from: "/app/$workspaceId/$channelId" });
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);
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

  useEffect(() => {
    setCurrentChannel(channelId);
  }, [channelId, setCurrentChannel]);

  return (
    <>
      <ChannelHeader workspaceId={workspaceId} channelId={channelId} />
      <div className="min-h-0 flex-1">
        <MessagePanel />
      </div>
      {/* モバイルでは入力欄の上に出す。デスクトップはサイドバーの下部 */}
      <MiniPlayer variant="mobile" />
      {isPreview ? (
        <JoinChannelBar
          workspaceId={workspaceId}
          channelId={channelId}
          channelName={channel.name}
        />
      ) : (
        <div className="relative">
          <TypingIndicator channelId={channelId} />
          <MessageInput key={channelId} channelId={channelId} />
        </div>
      )}
    </>
  );
};
