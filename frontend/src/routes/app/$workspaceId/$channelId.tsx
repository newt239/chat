import { useEffect } from "react";

import { createFileRoute } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { MessagePanel } from "@/features/message/components/MessagePanel";
import { setCurrentChannelAtom } from "@/providers/store/workspace";

const ChannelPage = () => {
  const { channelId } = Route.useParams();
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);

  useEffect(() => {
    if (channelId) {
      setCurrentChannel(channelId);
    }
  }, [channelId, setCurrentChannel]);

  return <MessagePanel />;
};

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelPage,
});
