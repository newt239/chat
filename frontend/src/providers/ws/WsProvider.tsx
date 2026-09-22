import { useEffect, useMemo, useState } from "react";

import { useAtomValue } from "jotai";

import { WsClient } from "#/lib/ws";
import { accessTokenAtom } from "#/providers/store/auth";
import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import { WsClientContext } from "./wsClientContext";

export const WsProvider = ({ children }: { children: React.ReactNode }) => {
  const accessToken = useAtomValue(accessTokenAtom);
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const [wsClient, setWsClient] = useState<WsClient | null>(null);

  useEffect(() => {
    if (!accessToken || !workspaceId) {
      setWsClient((prev) => {
        prev?.close();
        return null;
      });
      return undefined;
    }

    const instance = new WsClient(accessToken, workspaceId);
    setWsClient(instance);

    return () => {
      instance.close();
    };
  }, [accessToken, workspaceId]);

  const value = useMemo(() => ({ wsClient }), [wsClient]);
  return <WsClientContext.Provider value={value}>{children}</WsClientContext.Provider>;
};
