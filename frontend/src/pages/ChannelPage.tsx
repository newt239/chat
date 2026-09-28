import { useEffect } from "react";

import { useSetAtom } from "jotai";

import { MessagePanel } from "#/features/message/components/MessagePanel";
import { useChannelId } from "#/lib/routeParams";
import { setCurrentChannelAtom } from "#/providers/store/workspace";

export const ChannelPage = () => {
  const channelId = useChannelId();
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);

  useEffect(() => {
    setCurrentChannel(channelId);
  }, [channelId, setCurrentChannel]);

  return <MessagePanel />;
};
