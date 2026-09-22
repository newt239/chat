import { useContext } from "react";

import { WsClientContext } from "./wsClientContext";

export const useWsClient = () => useContext(WsClientContext);
