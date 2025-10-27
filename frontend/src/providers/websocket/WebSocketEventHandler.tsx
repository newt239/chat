import { useChannelEvents } from "./useChannelEvents";
import { useReadStateEvents } from "./useReadStateEvents";
import { useWebSocketEvents } from "./useWebSocketEvents";

export const WebSocketEventHandler = () => {
  // WebSocketイベントを処理
  useWebSocketEvents();

  // チャンネル参加・退出イベントを処理
  useChannelEvents();

  // 既読状態イベントを処理
  useReadStateEvents();

  return null;
};
