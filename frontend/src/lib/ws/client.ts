type WebSocketEventType =
  | "join_channel"
  | "leave_channel"
  | "post_message"
  | "typing"
  | "update_read_state"
  | "new_message"
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
  private url: string;
  private eventHandlers: Map<WebSocketEventType, EventHandler[]> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(workspaceId: string, accessToken: string) {
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
    this.url = `${wsUrl}/ws?workspaceId=${workspaceId}&token=${accessToken}`;
  }

  connect() {
    this.ws = new WebSocket(this.url);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketEvent = JSON.parse(event.data);
        this.handleEvent(message);
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    this.ws.onclose = () => {
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };
  }

  private attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      setTimeout(() => {
        this.connect();
      }, this.reconnectDelay * this.reconnectAttempts);
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
      this.ws.close();
      this.ws = null;
    }
  }
}
