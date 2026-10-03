import { useEffect } from "react";

import { useSetAtom } from "jotai";

import { channelViewersAtom } from "#/features/channel/atoms";
import { useWsClient } from "#/providers/ws/useWsClient";

/** ChannelViewers を受け取ってチャンネルごとの閲覧中のユーザーを置き換える。アプリの枠で一度だけ呼ぶ */
export const useChannelViewersSync = () => {
  const wsClient = useWsClient();
  const setViewers = useSetAtom(channelViewersAtom);

  useEffect(
    () =>
      wsClient?.on("channelViewers", ({ channelId, userIds }) => {
        setViewers((current) => ({ ...current, [channelId]: userIds }));
      }),
    [wsClient, setViewers],
  );
};

/** 画面に開いているチャンネルを閲覧中としてサーバーに伝える */
export const useViewChannel = (channelId: string) => {
  const wsClient = useWsClient();

  useEffect(() => {
    wsClient?.viewChannel(channelId);
    return () => {
      wsClient?.viewChannel("");
    };
  }, [wsClient, channelId]);
};
