import { createContext, useContext, useEffect, useMemo, useState } from "react";

import { useAtomValue } from "jotai";

import { WsClient } from "@/lib/ws";
import { accessTokenAtom } from "@/providers/store/auth";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

export type WsClientContextType = {
  wsClient: WsClient | null;
};

const WsClientContext = createContext<WsClientContextType>({ wsClient: null });

export const useWsClient = () => useContext(WsClientContext);

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
      return;
    }
    if (!wsClient) {
      setWsClient((prev) => {
        prev?.close();
        return null;
      });
      const instance = new WsClient(accessToken, workspaceId);
      setWsClient(instance);
    }

    return () => {
      setWsClient((prev) => {
        prev?.close();
        return null;
      });
    };
  }, [accessToken, workspaceId]);

  const value = useMemo(() => ({ wsClient }), [wsClient]);
  return <WsClientContext.Provider value={value}>{children}</WsClientContext.Provider>;
};
