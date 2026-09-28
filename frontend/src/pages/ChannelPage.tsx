import { useEffect } from "react";

import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ChannelHeader } from "#/features/channel/components/ChannelHeader";
import { MessageInput } from "#/features/message/components/MessageInput";
import { MessagePanel } from "#/features/message/components/MessagePanel";
import { setCurrentChannelAtom } from "#/providers/store/workspace";

export const ChannelPage = () => {
  const { workspaceId, channelId } = useParams({ from: "/app/$workspaceId/$channelId" });
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);

  useEffect(() => {
    setCurrentChannel(channelId);
  }, [channelId, setCurrentChannel]);

  return (
    <>
      <ChannelHeader workspaceId={workspaceId} channelId={channelId} />
      <div className="min-h-0 flex-1">
        <MessagePanel />
      </div>
      <MessageInput channelId={channelId} />
    </>
  );
};
