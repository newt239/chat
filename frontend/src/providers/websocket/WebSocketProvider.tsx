import React, { createContext, useContext, useEffect, useRef, useState } from "react";

import { useAtomValue } from "jotai";

import { WebSocketClient } from "./client";

import { accessTokenAtom } from "@/providers/store/auth";
import { currentWorkspaceIdAtom } from "@/providers/store/workspace";

type WebSocketContextType = {
  client: WebSocketClient | null;
  isConnected: boolean;
};

const WebSocketContext = createContext<WebSocketContextType>({
  client: null,
  isConnected: false,
});

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};

type WebSocketProviderProps = {
  children: React.ReactNode;
};

export const WebSocketProvider = ({ children }: WebSocketProviderProps) => {
  const accessToken = useAtomValue(accessTokenAtom);
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const clientRef = useRef<WebSocketClient | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const isConnectingRef = useRef(false);
  const shouldReconnectRef = useRef(true);

  useEffect(() => {
    if (!accessToken || !workspaceId) {
      // 認証情報がない場合は接続を切断
      if (clientRef.current) {
        clientRef.current.disconnect();
        clientRef.current = null;
        setIsConnected(false);
      }
      isConnectingRef.current = false;
      return;
    }

    // アクセストークンが有効かチェック
    if (accessToken.length === 0) {
      console.warn("Access token is empty, skipping WebSocket connection");
      return;
    }

    // 既に接続中または接続済みの場合は何もしない
    if (
      isConnectingRef.current ||
      (clientRef.current && clientRef.current.webSocket?.readyState === WebSocket.OPEN)
    ) {
      return;
    }

    isConnectingRef.current = true;

    // 既存の接続がある場合は切断
    if (clientRef.current) {
      clientRef.current.disconnect();
    }

    // 新しいWebSocket接続を作成
    const client = new WebSocketClient(workspaceId, accessToken);
    clientRef.current = client;

    // 接続状態を監視するためのコールバックを設定
    client.setConnectionCallbacks(
      () => {
        console.log("WebSocket connected successfully");
        setIsConnected(true);
        isConnectingRef.current = false;
      }, // onOpen
      () => {
        console.log("WebSocket disconnected");
        setIsConnected(false);
        isConnectingRef.current = false;
      }, // onClose
      () => {
        console.log("WebSocket error occurred");
        setIsConnected(false);
        isConnectingRef.current = false;
      } // onError
    );

    // 接続を開始
    client.connect();

    return () => {
      console.log("WebSocketProvider cleanup");
      shouldReconnectRef.current = false;
      isConnectingRef.current = false;
      if (clientRef.current) {
        clientRef.current.disconnect();
        clientRef.current = null;
        setIsConnected(false);
      }
    };
  }, [accessToken, workspaceId]);

  const value: WebSocketContextType = {
    client: clientRef.current,
    isConnected,
  };

  return <WebSocketContext.Provider value={value}>{children}</WebSocketContext.Provider>;
};
