import { createContext, useContext } from "react";

import type { WsClient } from "#/lib/ws";

export const WsClientContext = createContext<WsClient | null>(null);

export const useWsClient = () => useContext(WsClientContext);
