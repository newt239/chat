import { useEffect } from "react";

import { createFileRoute } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { MessagePanel } from "@/features/message/components/MessagePanel";
import { setIsChannelPageAtom } from "@/providers/store/ui";
import { setCurrentChannelAtom } from "@/providers/store/workspace";

const ChannelPage = () => {
  const { channelId } = Route.useParams();
  const setCurrentChannel = useSetAtom(setCurrentChannelAtom);
  const setIsChannelPage = useSetAtom(setIsChannelPageAtom);

  useEffect(() => {
    if (channelId) {
      setCurrentChannel(channelId);
    }
    // チャンネルページであることを設定
    setIsChannelPage(true);

    // コンポーネントのアンマウント時にチャンネルページでないことを設定
    return () => {
      setIsChannelPage(false);
    };
  }, [channelId, setCurrentChannel, setIsChannelPage]);

  return <MessagePanel />;
};

export const Route = createFileRoute("/app/$workspaceId/$channelId")({
  component: ChannelPage,
});
