import { useEffect, useMemo, useState } from "react";

import { createClient } from "@connectrpc/connect";
import { useAtomValue } from "jotai";

import { RealtimeService } from "#/gen/chat/v1/realtime_service_pb";
import { transport } from "#/lib/api/transport";
import { getIsMobileApp, isTauri } from "#/lib/platform/platform";
import { WsClient } from "#/lib/ws";
import { sessionAtom } from "#/providers/store/auth";

import { WsClientContext } from "./wsClientContext";

const realtimeClient = createClient(RealtimeService, transport);

type WsProviderProps = {
  workspaceId: string;
  children: React.ReactNode;
};

export const WsProvider = ({ workspaceId, children }: WsProviderProps) => {
  // トークンの更新では接続し直さないよう、セッションの有無だけを見る
  const hasSession = useAtomValue(sessionAtom) !== null;
  const [wsClient, setWsClient] = useState<WsClient | null>(null);

  useEffect(() => {
    if (!hasSession) {
      setWsClient(null);
      return undefined;
    }
    const instance = new WsClient(
      () => realtimeClient.issueWebSocketTicket({ workspaceId }).then(({ ticket }) => ticket),
      isTauri && !getIsMobileApp(),
    );
    setWsClient(instance);
    return () => {
      instance.close();
    };
  }, [hasSession, workspaceId]);

  const value = useMemo(() => ({ wsClient }), [wsClient]);
  return <WsClientContext.Provider value={value}>{children}</WsClientContext.Provider>;
};
