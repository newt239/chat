import { create, fromJsonString, toJsonString } from "@bufbuild/protobuf";

import { ClientEventSchema, ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { logger } from "#/lib/logger";
import { navigateTo } from "#/lib/navigation";

import type { ServerEvent } from "#/gen/chat/v1/event_pb";

import type { MessageInitShape } from "@bufbuild/protobuf";

type ServerEventOneof = Exclude<ServerEvent["event"], { case: undefined }>;
type WsEventType = ServerEventOneof["case"];
type WsEventPayload<K extends WsEventType> = Extract<ServerEventOneof, { case: K }>["value"];

const WS_BC_NAME = "ws-control";
const WS_RECONNECT_DELAY = 2_000; // 初期遅延: 2秒
const WS_MAX_RECONNECT_DELAY = 30_000; // 最大遅延: 30秒
const WS_MAX_RECONNECT_ATTEMPTS = 5; // 最大再接続試行回数

/** サーバWebSocketエンドポイント取得 例: ws://localhost:8080/ws?token=xxxx&workspaceId=xxxx */
const getWsUrl = (token: string, workspaceId: string): string => {
  const base = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080";
  return `${base}/ws?token=${encodeURIComponent(token)}&workspaceId=${encodeURIComponent(workspaceId)}`;
};

const parseServerEvent = (data: string) => {
  try {
    return fromJsonString(ServerEventSchema, data, { ignoreUnknownFields: true });
  } catch (error) {
    logger.warn("WebSocketイベントの形式が想定と異なります:", error);
    return null;
  }
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

  private readonly handlers: {
    [K in WsEventType]: Set<(payload: WsEventPayload<K>) => void>;
  } = {
    ack: new Set(),
    channelViewers: new Set(),
    error: new Set(),
    messageDeleted: new Set(),
    messageUpdated: new Set(),
    newMessage: new Set(),
    pinCreated: new Set(),
    pinDeleted: new Set(),
    reactionAdded: new Set(),
    reactionRemoved: new Set(),
    stopTyping: new Set(),
    systemMessageCreated: new Set(),
    typing: new Set(),
    unreadCount: new Set(),
  };

  public constructor(token: string, workspaceId: string) {
    this.token = token;
    this.workspaceId = workspaceId;
    this.bc = new BroadcastChannel(WS_BC_NAME);
    this.listenBroadcast();
    this.initTabActivityControl();
  }

  /** サーバーイベントを購読中のハンドラへ配る（WebSocket の message ハンドラ） */
  public readonly eventDispatcher = (event: MessageEvent<string>) => {
    const oneof = parseServerEvent(event.data)?.event;
    if (oneof?.case === undefined) {
      return;
    }

    // payload の型をイベントの種類ごとに絞り込むため、case ごとに emit する
    switch (oneof.case) {
      case "newMessage": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "messageUpdated": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "messageDeleted": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "unreadCount": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "pinCreated": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "pinDeleted": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "systemMessageCreated": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "reactionAdded": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "reactionRemoved": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "typing": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "stopTyping": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "ack": {
        this.emit(oneof.case, oneof.value);
        break;
      }
      case "error": {
        if (oneof.value.code === "401") {
          navigateTo({ to: "/login" });
        }
        this.emit(oneof.case, oneof.value);
        break;
      }
      default: {
        break;
      }
    }
  };

  private emit<T extends WsEventType>(type: T, payload: WsEventPayload<T>) {
    for (const handler of this.handlers[type]) {
      handler(payload);
    }
  }

  /** サーバーイベントの購読を開始する。戻り値を呼ぶと購読を解除する */
  public on<T extends WsEventType>(type: T, cb: (payload: WsEventPayload<T>) => void) {
    this.handlers[type].add(cb);

    return () => {
      this.off(type, cb);
    };
  }

  /** サーバーイベントの購読を解除する */
  public off<T extends WsEventType>(type: T, cb: (payload: WsEventPayload<T>) => void) {
    this.handlers[type].delete(cb);
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

  public joinChannel(channelId: string) {
    this.send({ case: "joinChannel", value: { channelId } });
  }
  public leaveChannel(channelId: string) {
    this.send({ case: "leaveChannel", value: { channelId } });
  }
  public typing(channelId: string) {
    this.send({ case: "typing", value: { channelId } });
  }
  public stopTyping(channelId: string) {
    this.send({ case: "stopTyping", value: { channelId } });
  }

  private send(event: NonNullable<MessageInitShape<typeof ClientEventSchema>["event"]>) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(toJsonString(ClientEventSchema, create(ClientEventSchema, { event })));
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
