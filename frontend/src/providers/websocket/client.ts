type WebSocketEventType =
  | "join_channel"
  | "leave_channel"
  | "post_message"
  | "typing"
  | "update_read_state"
  | "new_message"
  | "message_updated"
  | "message_deleted"
  | "unread_count"
  | "ack"
  | "error";

type WebSocketEvent = {
  type: WebSocketEventType;
  payload?: unknown;
};

type EventHandler = (payload: unknown) => void;

export class WebSocketClient {
  private ws: WebSocket | null = null;
  private workspaceId: string;
  private accessToken: string;
  private eventHandlers: Map<WebSocketEventType, EventHandler[]> = new Map();
  private onOpenCallback?: () => void;
  private onCloseCallback?: () => void;
  private onErrorCallback?: () => void;

  get webSocket() {
    return this.ws;
  }

  constructor(workspaceId: string, accessToken: string) {
    this.workspaceId = workspaceId;
    this.accessToken = accessToken;
  }

  connect() {
    // 既存の接続がある場合は切断
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
    // WebSocketはブラウザAPIのためAuthorizationヘッダーを直接設定できない
    // クエリパラメータでトークンとWorkspaceIDを送信
    const url = `${wsUrl}/ws?workspaceId=${this.workspaceId}&token=${this.accessToken}`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        console.log("WebSocket connected successfully");
        this.onOpenCallback?.();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketEvent = JSON.parse(event.data);
          this.handleEvent(message);
        } catch (error) {
          console.error("Failed to parse WebSocket message:", error);
        }
      };

      this.ws.onclose = (event) => {
        console.log("WebSocket closed:", event.code, event.reason);
        this.onCloseCallback?.();
        // 再接続はWebSocketProviderで制御するため、ここでは行わない
      };

      this.ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        console.error("WebSocket URL:", url);
        console.error("Workspace ID:", this.workspaceId);
        console.error("Access Token:", this.accessToken ? "Present" : "Missing");
        this.onErrorCallback?.();
      };
    } catch (error) {
      console.error("Failed to create WebSocket connection:", error);
      this.onErrorCallback?.();
    }
  }

  private handleEvent(message: WebSocketEvent) {
    const handlers = this.eventHandlers.get(message.type);
    if (handlers) {
      handlers.forEach((handler) => handler(message.payload));
    }
  }

  on(eventType: WebSocketEventType, handler: EventHandler) {
    const handlers = this.eventHandlers.get(eventType) || [];
    handlers.push(handler);
    this.eventHandlers.set(eventType, handlers);
  }

  off(eventType: WebSocketEventType, handler: EventHandler) {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index !== -1) {
        handlers.splice(index, 1);
      }
    }
  }

  send(type: WebSocketEventType, payload?: unknown) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }

  disconnect() {
    if (this.ws) {
      // 正常な切断コード（1000）で閉じる
      this.ws.close(1000, "Client disconnect");
      this.ws = null;
    }
  }

  setConnectionCallbacks(onOpen?: () => void, onClose?: () => void, onError?: () => void) {
    this.onOpenCallback = onOpen;
    this.onCloseCallback = onClose;
    this.onErrorCallback = onError;
  }
}
