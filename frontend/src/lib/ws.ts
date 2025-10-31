import type { ClientToServerMessage, WsEventPayloadMap } from "@/types/wsEvents";

import { router } from "@/lib/router";

const WS_BC_NAME = "ws-control";

/**
 * サーバWebSocketエンドポイント取得
 * 例: ws://localhost:8080/ws?token=xxxx&workspaceId=xxxx
 */
function getWsUrl(token: string, workspaceId: string): string {
  const base = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
  return `${base}/ws?token=${encodeURIComponent(token)}&workspaceId=${encodeURIComponent(workspaceId)}`;
}

export class WsClient {
  private ws: WebSocket | null = null;
  private heartbeatInterval: number = 30000; // 30秒
  private pingIntervalId: ReturnType<typeof setInterval> | null = null;
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = 2000; // ms

  private token: string;
  private workspaceId: string;
  private bc: BroadcastChannel;
  private isActiveLeader: boolean = false;
  // 各イベント専用callback配列で厳密管理
  private handlers = {
    new_message: [] as ((payload: WsEventPayloadMap["new_message"]) => void)[],
    message_updated: [] as ((payload: WsEventPayloadMap["message_updated"]) => void)[],
    message_deleted: [] as ((payload: WsEventPayloadMap["message_deleted"]) => void)[],
    unread_count: [] as ((payload: WsEventPayloadMap["unread_count"]) => void)[],
    pin_created: [] as ((payload: WsEventPayloadMap["pin_created"]) => void)[],
    pin_deleted: [] as ((payload: WsEventPayloadMap["pin_deleted"]) => void)[],
    system_message_created: [] as ((payload: WsEventPayloadMap["system_message_created"]) => void)[],
    ack: [] as ((payload: WsEventPayloadMap["ack"]) => void)[],
    error: [] as ((payload: WsEventPayloadMap["error"]) => void)[],
  };

  constructor(token: string, workspaceId: string) {
    this.token = token;
    this.workspaceId = workspaceId;
    this.bc = new BroadcastChannel(WS_BC_NAME);
    this.listenBroadcast();
    this.initTabActivityControl();
  }

  // JSON文字列しか来ない前提で型安全に
  private eventDispatcher = (event: MessageEvent<string>) => {
    try {
      type EventUnion = {
        [K in keyof WsEventPayloadMap]: { type: K; payload: WsEventPayloadMap[K] };
      }[keyof WsEventPayloadMap];
      const parsed = JSON.parse(event.data) as EventUnion;
      if (!parsed || typeof parsed.type !== "string") return;
      const { type, payload } = parsed;
      switch (type) {
        case "new_message":
          this.handlers.new_message.forEach((cb) => cb(payload));
          break;
        case "message_updated":
          this.handlers.message_updated.forEach((cb) => cb(payload));
          break;
        case "message_deleted":
          this.handlers.message_deleted.forEach((cb) => cb(payload));
          break;
        case "unread_count":
          this.handlers.unread_count.forEach((cb) => cb(payload));
          break;
        case "pin_created":
          this.handlers.pin_created.forEach((cb) => cb(payload));
          break;
        case "pin_deleted":
          this.handlers.pin_deleted.forEach((cb) => cb(payload));
          break;
        case "system_message_created":
          this.handlers.system_message_created.forEach((cb) => cb(payload));
          break;
        case "ack":
          this.handlers.ack.forEach((cb) => cb(payload));
          break;
        case "error":
          if (typeof payload === "object" && "code" in payload && payload.code === "401") {
            router.navigate({ to: "/login" });
          }
          this.handlers.error.forEach((cb) => cb(payload));
          break;
      }
    } catch {
      // nop
    }
  };

  public onNewMessage(cb: (payload: WsEventPayloadMap["new_message"]) => void) {
    this.handlers.new_message.push(cb);
  }
  public onMessageUpdated(cb: (payload: WsEventPayloadMap["message_updated"]) => void) {
    this.handlers.message_updated.push(cb);
  }
  public onMessageDeleted(cb: (payload: WsEventPayloadMap["message_deleted"]) => void) {
    this.handlers.message_deleted.push(cb);
  }
  public onUnreadCount(cb: (payload: WsEventPayloadMap["unread_count"]) => void) {
    this.handlers.unread_count.push(cb);
  }
  public onPinCreated(cb: (payload: WsEventPayloadMap["pin_created"]) => void) {
    this.handlers.pin_created.push(cb);
  }
  public onPinDeleted(cb: (payload: WsEventPayloadMap["pin_deleted"]) => void) {
    this.handlers.pin_deleted.push(cb);
  }
  public onSystemMessageCreated(cb: (payload: WsEventPayloadMap["system_message_created"]) => void) {
    this.handlers.system_message_created.push(cb);
  }
  public onAck(cb: (payload: WsEventPayloadMap["ack"]) => void) {
    this.handlers.ack.push(cb);
  }
  public onWsError(cb: (payload: WsEventPayloadMap["error"]) => void) {
    this.handlers.error.push(cb);
  }

  private connect() {
    const url = getWsUrl(this.token, this.workspaceId);
    this.ws = new WebSocket(url);
    this.ws.addEventListener("open", this.onOpen);
    this.ws.addEventListener("close", this.onClose);
    this.ws.addEventListener("error", this.onError);
    this.ws.addEventListener("message", this.eventDispatcher);
  }

  private onOpen = () => {
    this.startHeartbeat();
    console.log("WebSocket接続が開きました", this.workspaceId);
  };

  private startHeartbeat() {
    this.pingIntervalId = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: "ping" }));
      }
    }, this.heartbeatInterval);
  }

  private stopHeartbeat() {
    if (this.pingIntervalId) clearInterval(this.pingIntervalId);
    this.pingIntervalId = null;
  }

  private onClose = () => {
    this.stopHeartbeat();
    this.tryReconnect();
    console.log("WebSocket接続が閉じました", this.workspaceId);
  };

  private onError = () => {
    this.close();
    this.tryReconnect();
  };

  private tryReconnect() {
    if (this.reconnectTimeoutId) return;
    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null;
      this.connect();
    }, this.reconnectDelay);
  }

  public joinChannel(channel_id: string) {
    this.send({ type: "join_channel", payload: { channel_id } });
  }
  public leaveChannel(channel_id: string) {
    this.send({ type: "leave_channel", payload: { channel_id } });
  }
  public postMessage(channel_id: string, body: string) {
    this.send({ type: "post_message", payload: { channel_id, body } });
  }
  public typing(channel_id: string) {
    this.send({ type: "typing", payload: { channel_id } });
  }
  public updateReadState(channel_id: string, message_id: string) {
    this.send({ type: "update_read_state", payload: { channel_id, message_id } });
  }

  private send(data: ClientToServerMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  public close() {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.removeEventListener("open", this.onOpen);
      this.ws.removeEventListener("close", this.onClose);
      this.ws.removeEventListener("error", this.onError);
      this.ws.removeEventListener("message", this.eventDispatcher);
      this.ws.close();
      this.ws = null;
    }
    if (this.reconnectTimeoutId) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }
  }

  private listenBroadcast() {
    this.bc.onmessage = (ev) => {
      if (!ev.data) return;
      // 他タブが接続を開始したら自分はリーダー権を放棄
      if (ev.data.type === "ws_active") {
        this.isActiveLeader = false;
        this.close();
      }
    };
  }

  private initTabActivityControl() {
    window.addEventListener("visibilitychange", this.handleVisibility, false);
    window.addEventListener("focus", this.handleFocus, false);
    window.addEventListener("beforeunload", this.handleUnload, false);
    // 初回ロード判定
    setTimeout(() => {
      if (document.visibilityState === "visible") {
        this.becomeLeaderAndConnect();
      }
    }, 0);
  }

  private handleVisibility = () => {
    if (document.visibilityState === "visible") {
      this.becomeLeaderAndConnect();
    } else {
      this.isActiveLeader = false;
      this.close();
    }
  };

  private handleFocus = () => {
    if (!this.isActiveLeader) {
      this.becomeLeaderAndConnect();
    }
  };

  private handleUnload = () => {
    this.isActiveLeader = false;
    this.close();
    this.bc.close();
  };

  private becomeLeaderAndConnect() {
    if (!this.isActiveLeader) {
      this.isActiveLeader = true;
      this.bc.postMessage({ type: "ws_active" });
      this.connect();
    }
  }
}
