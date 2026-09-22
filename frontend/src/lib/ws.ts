import { logger } from "#/lib/logger";
import { navigateTo } from "#/lib/navigation";
import { paths } from "#/lib/paths";
import { parseServerEvent } from "#/types/wsEvents";

import type { ClientToServerMessage, WsEventPayloadMap } from "#/types/wsEvents";

const WS_BC_NAME = "ws-control";
const WS_RECONNECT_DELAY = 2_000; // 初期遅延: 2秒
const WS_MAX_RECONNECT_DELAY = 30_000; // 最大遅延: 30秒
const WS_MAX_RECONNECT_ATTEMPTS = 5; // 最大再接続試行回数

/** サーバWebSocketエンドポイント取得 例: ws://localhost:8080/ws?token=xxxx&workspaceId=xxxx */
const getWsUrl = (token: string, workspaceId: string): string => {
  const base = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080";
  return `${base}/ws?token=${encodeURIComponent(token)}&workspaceId=${encodeURIComponent(workspaceId)}`;
};

export class WsClient {
  private ws: WebSocket | null = null;
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private reconnectDelay = WS_RECONNECT_DELAY;
  private reconnectAttempts = 0;
  private shouldStopReconnecting = false;

  private readonly token: string;
  private readonly workspaceId: string;
  private readonly bc: BroadcastChannel;
  private isActiveLeader = false;

  private readonly handlers = {
    ack: [] as ((payload: WsEventPayloadMap["ack"]) => void)[],
    error: [] as ((payload: WsEventPayloadMap["error"]) => void)[],
    message_deleted: [] as ((payload: WsEventPayloadMap["message_deleted"]) => void)[],
    message_updated: [] as ((payload: WsEventPayloadMap["message_updated"]) => void)[],
    new_message: [] as ((payload: WsEventPayloadMap["new_message"]) => void)[],
    pin_created: [] as ((payload: WsEventPayloadMap["pin_created"]) => void)[],
    pin_deleted: [] as ((payload: WsEventPayloadMap["pin_deleted"]) => void)[],
    system_message_created: [] as ((
      payload: WsEventPayloadMap["system_message_created"],
    ) => void)[],
    unread_count: [] as ((payload: WsEventPayloadMap["unread_count"]) => void)[],
  };

  public constructor(token: string, workspaceId: string) {
    this.token = token;
    this.workspaceId = workspaceId;
    this.bc = new BroadcastChannel(WS_BC_NAME);
    this.listenBroadcast();
    this.initTabActivityControl();
  }

  private readonly eventDispatcher = (event: MessageEvent<string>) => {
    try {
      const parsed = parseServerEvent(event.data);
      if (!parsed.success) {
        logger.warn("WebSocketイベントの形式が想定と異なります:", parsed.error);
        return;
      }
      const { type, payload } = parsed.data;
      switch (type) {
        case "new_message": {
          for (const cb of this.handlers.new_message) {
            cb(payload);
          }
          break;
        }
        case "message_updated": {
          for (const cb of this.handlers.message_updated) {
            cb(payload);
          }
          break;
        }
        case "message_deleted": {
          for (const cb of this.handlers.message_deleted) {
            cb(payload);
          }
          break;
        }
        case "unread_count": {
          for (const cb of this.handlers.unread_count) {
            cb(payload);
          }
          break;
        }
        case "pin_created": {
          for (const cb of this.handlers.pin_created) {
            cb(payload);
          }
          break;
        }
        case "pin_deleted": {
          for (const cb of this.handlers.pin_deleted) {
            cb(payload);
          }
          break;
        }
        case "system_message_created": {
          for (const cb of this.handlers.system_message_created) {
            cb(payload);
          }
          break;
        }
        case "ack": {
          for (const cb of this.handlers.ack) {
            cb(payload);
          }
          break;
        }
        case "error": {
          if (typeof payload === "object" && "code" in payload && payload.code === "401") {
            navigateTo(paths.login());
          }
          for (const cb of this.handlers.error) {
            cb(payload);
          }
          break;
        }
        default: {
          break;
        }
      }
    } catch (error) {
      logger.error("WebSocketイベント処理エラー:", error);
    }
  };

  public onNewMessage(cb: (payload: WsEventPayloadMap["new_message"]) => void) {
    this.handlers.new_message.push(cb);
  }
  public offNewMessage(cb: (payload: WsEventPayloadMap["new_message"]) => void) {
    const index = this.handlers.new_message.indexOf(cb);
    if (index !== -1) {
      this.handlers.new_message.splice(index, 1);
    }
  }
  public onMessageUpdated(cb: (payload: WsEventPayloadMap["message_updated"]) => void) {
    this.handlers.message_updated.push(cb);
  }
  public offMessageUpdated(cb: (payload: WsEventPayloadMap["message_updated"]) => void) {
    const index = this.handlers.message_updated.indexOf(cb);
    if (index !== -1) {
      this.handlers.message_updated.splice(index, 1);
    }
  }
  public onMessageDeleted(cb: (payload: WsEventPayloadMap["message_deleted"]) => void) {
    this.handlers.message_deleted.push(cb);
  }
  public offMessageDeleted(cb: (payload: WsEventPayloadMap["message_deleted"]) => void) {
    const index = this.handlers.message_deleted.indexOf(cb);
    if (index !== -1) {
      this.handlers.message_deleted.splice(index, 1);
    }
  }
  public onUnreadCount(cb: (payload: WsEventPayloadMap["unread_count"]) => void) {
    this.handlers.unread_count.push(cb);
  }
  public offUnreadCount(cb: (payload: WsEventPayloadMap["unread_count"]) => void) {
    const index = this.handlers.unread_count.indexOf(cb);
    if (index !== -1) {
      this.handlers.unread_count.splice(index, 1);
    }
  }
  public onPinCreated(cb: (payload: WsEventPayloadMap["pin_created"]) => void) {
    this.handlers.pin_created.push(cb);
  }
  public offPinCreated(cb: (payload: WsEventPayloadMap["pin_created"]) => void) {
    const index = this.handlers.pin_created.indexOf(cb);
    if (index !== -1) {
      this.handlers.pin_created.splice(index, 1);
    }
  }
  public onPinDeleted(cb: (payload: WsEventPayloadMap["pin_deleted"]) => void) {
    this.handlers.pin_deleted.push(cb);
  }
  public offPinDeleted(cb: (payload: WsEventPayloadMap["pin_deleted"]) => void) {
    const index = this.handlers.pin_deleted.indexOf(cb);
    if (index !== -1) {
      this.handlers.pin_deleted.splice(index, 1);
    }
  }
  public onSystemMessageCreated(
    cb: (payload: WsEventPayloadMap["system_message_created"]) => void,
  ) {
    this.handlers.system_message_created.push(cb);
  }
  public offSystemMessageCreated(
    cb: (payload: WsEventPayloadMap["system_message_created"]) => void,
  ) {
    const index = this.handlers.system_message_created.indexOf(cb);
    if (index !== -1) {
      this.handlers.system_message_created.splice(index, 1);
    }
  }
  public onAck(cb: (payload: WsEventPayloadMap["ack"]) => void) {
    this.handlers.ack.push(cb);
  }
  public offAck(cb: (payload: WsEventPayloadMap["ack"]) => void) {
    const index = this.handlers.ack.indexOf(cb);
    if (index !== -1) {
      this.handlers.ack.splice(index, 1);
    }
  }
  public onWsError(cb: (payload: WsEventPayloadMap["error"]) => void) {
    this.handlers.error.push(cb);
  }
  public offWsError(cb: (payload: WsEventPayloadMap["error"]) => void) {
    const index = this.handlers.error.indexOf(cb);
    if (index !== -1) {
      this.handlers.error.splice(index, 1);
    }
  }

  private connect() {
    if (this.shouldStopReconnecting) {
      logger.info("WebSocket再接続を停止しました", this.workspaceId);
      return;
    }

    const url = getWsUrl(this.token, this.workspaceId);
    logger.info("WebSocket接続開始:", url);
    try {
      this.ws = new WebSocket(url);
      this.ws.addEventListener("open", this.onOpen);
      this.ws.addEventListener("close", this.onClose);
      this.ws.addEventListener("error", this.onError);
      this.ws.addEventListener("message", this.eventDispatcher);
    } catch (error) {
      logger.error("WebSocket接続作成時エラー:", error);
      this.handleConnectionFailure("接続作成エラー", error);
    }
  }

  private readonly onOpen = () => {
    logger.info("WebSocket接続が開きました", this.workspaceId);
    // 接続成功時は再接続試行回数をリセット
    this.reconnectAttempts = 0;
    this.reconnectDelay = WS_RECONNECT_DELAY;
    this.shouldStopReconnecting = false;
  };

  private readonly onClose = (event: CloseEvent) => {
    logger.info("WebSocket接続が閉じました", {
      code: event.code,
      reason: event.reason,
      wasClean: event.wasClean,
      workspaceId: this.workspaceId,
    });
    // リーダーの場合のみ再接続を試みる
    if (this.isActiveLeader && !this.shouldStopReconnecting) {
      // 正常終了（1000）の場合は再接続を試みない
      if (event.code === 1000) {
        logger.info("WebSocket正常終了のため再接続しません", this.workspaceId);
        return;
      }
      // 認証エラー（1008）の場合は再接続を停止
      if (event.code === 1008) {
        logger.error("WebSocket認証エラーのため再接続を停止します", this.workspaceId);
        this.shouldStopReconnecting = true;
        return;
      }
      this.handleConnectionFailure("接続が閉じられました", event);
    }
  };

  private readonly onError = (event: Event) => {
    const errorInfo = this.getErrorInfo(event);
    logger.error("WebSocketエラーが発生しました", {
      error: errorInfo,
      readyState: this.ws?.readyState,
      workspaceId: this.workspaceId,
    });
    // リーダーの場合のみ再接続を試みる
    if (this.isActiveLeader && !this.shouldStopReconnecting) {
      this.handleConnectionFailure("WebSocketエラー", event);
    }
  };

  private handleConnectionFailure(context: string, error: unknown) {
    if (this.shouldStopReconnecting) {
      return;
    }

    this.reconnectAttempts += 1;

    if (this.reconnectAttempts >= WS_MAX_RECONNECT_ATTEMPTS) {
      logger.error("WebSocket最大再接続試行回数に達しました", {
        attempts: this.reconnectAttempts,
        context,
        error: this.getErrorInfo(error),
        workspaceId: this.workspaceId,
      });
      this.shouldStopReconnecting = true;
      return;
    }

    logger.info("WebSocket再接続を試みます", {
      attempt: this.reconnectAttempts,
      context,
      delay: this.reconnectDelay,
      error: this.getErrorInfo(error),
      maxAttempts: WS_MAX_RECONNECT_ATTEMPTS,
      workspaceId: this.workspaceId,
    });

    this.tryReconnect();
  }

  private tryReconnect() {
    if (this.reconnectTimeoutId) {
      return;
    }
    if (!this.isActiveLeader) {
      return;
    }
    if (this.shouldStopReconnecting) {
      return;
    }

    this.reconnectTimeoutId = setTimeout(() => {
      this.reconnectTimeoutId = null;
      if (this.isActiveLeader && !this.shouldStopReconnecting) {
        // 指数バックオフ: 2秒, 4秒, 8秒, 16秒, 最大30秒
        this.reconnectDelay = Math.min(
          WS_RECONNECT_DELAY * 2 ** (this.reconnectAttempts - 1),
          WS_MAX_RECONNECT_DELAY,
        );
        this.connect();
      }
    }, this.reconnectDelay);
  }

  private getErrorInfo(event: unknown): string {
    if (event instanceof ErrorEvent) {
      return event.message;
    }
    if (event instanceof CloseEvent) {
      return `CloseEvent: code=${event.code}, reason=${event.reason}`;
    }
    if (event instanceof Error) {
      return event.message;
    }
    return JSON.stringify(event);
  }

  public joinChannel(channel_id: string) {
    this.send({ payload: { channel_id }, type: "join_channel" });
  }
  public leaveChannel(channel_id: string) {
    this.send({ payload: { channel_id }, type: "leave_channel" });
  }
  public postMessage(channel_id: string, body: string) {
    this.send({ payload: { body, channel_id }, type: "post_message" });
  }
  public typing(channel_id: string) {
    this.send({ payload: { channel_id }, type: "typing" });
  }
  public updateReadState(channel_id: string, message_id: string) {
    this.send({ payload: { channel_id, message_id }, type: "update_read_state" });
  }

  private send(data: ClientToServerMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  public close() {
    this.isActiveLeader = false;
    this.shouldStopReconnecting = true;
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
    window.removeEventListener("visibilitychange", this.handleVisibility, false);
    window.removeEventListener("focus", this.handleFocus, false);
    window.removeEventListener("beforeunload", this.handleUnload, false);
    this.bc.close();
    // 再接続状態をリセット
    this.reconnectAttempts = 0;
    this.reconnectDelay = WS_RECONNECT_DELAY;
  }

  private listenBroadcast() {
    // 他タブが接続を開始したら自分はリーダー権を放棄
    this.bc.addEventListener("message", () => {
      this.isActiveLeader = false;
      this.close();
    });
  }

  private initTabActivityControl() {
    window.addEventListener("visibilitychange", this.handleVisibility, false);
    window.addEventListener("focus", this.handleFocus, false);
    window.addEventListener("beforeunload", this.handleUnload, false);
    // 初回ロード時、ページが可視状態であれば接続
    if (document.visibilityState === "visible" && document.hasFocus()) {
      this.becomeLeaderAndConnect();
    }
  }

  private readonly handleVisibility = () => {
    if (document.visibilityState === "visible") {
      this.becomeLeaderAndConnect();
    } else {
      this.isActiveLeader = false;
      this.close();
    }
  };

  private readonly handleFocus = () => {
    if (!this.isActiveLeader) {
      this.becomeLeaderAndConnect();
    }
  };

  private readonly handleUnload = () => {
    this.isActiveLeader = false;
    this.close();
    this.bc.close();
  };

  private becomeLeaderAndConnect() {
    if (!this.isActiveLeader) {
      this.isActiveLeader = true;
      this.shouldStopReconnecting = false;
      this.reconnectAttempts = 0;
      this.reconnectDelay = WS_RECONNECT_DELAY;
      this.bc.postMessage({ type: "ws_active" });
      this.connect();
    }
  }
}
