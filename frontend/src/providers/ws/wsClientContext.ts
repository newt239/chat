import { createContext } from "react";

import type { WsClient } from "#/lib/ws";

export type WsClientContextValue = {
  wsClient: WsClient | null;
};

export const WsClientContext = createContext<WsClientContextValue>({ wsClient: null });
