import { useEffect } from "react";

import { createFileRoute } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { ChatLayout } from "@/features/message/components/ChatLayout";
import { setCurrentChannelAtom } from "@/lib/store/workspace";

const ChannelComponent = () => {
  const { workspaceId, channelId } = Route.useParams();
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);

  useEffect(() => {
    setCurrentChannel(channelId);
  }, [channelId, setCurrentChannel]);

  return <ChatLayout workspaceId={workspaceId} channelId={channelId} />;
};

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelComponent,
});
