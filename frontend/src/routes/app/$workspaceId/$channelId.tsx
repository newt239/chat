import { useEffect } from "react";

import { createFileRoute } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { MessagePanel } from "@/features/message/components/MessagePanel";
import { setCurrentChannelAtom } from "@/lib/store/workspace";

const ChannelComponent = () => {
  const { workspaceId, channelId } = Route.useParams();
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);

  useEffect(() => {
    setCurrentChannel(channelId);
  }, [channelId, setCurrentChannel]);

  return <MessagePanel workspaceId={workspaceId} channelId={channelId} />;
};

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelComponent,
});
