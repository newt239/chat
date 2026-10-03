import { create, fromJsonString, toJsonString } from "@bufbuild/protobuf";

import { ClientEventSchema, ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { navigateTo } from "#/lib/navigation";
import { refreshOrSignOut } from "#/lib/session";

import type { ServerEvent } from "#/gen/chat/v1/event_pb";

import type { MessageInitShape } from "@bufbuild/protobuf";

type ServerEventOneof = Exclude<ServerEvent["event"], { case: undefined }>;
type WsEventType = ServerEventOneof["case"];
type WsEventPayload<K extends WsEventType> = Extract<ServerEventOneof, { case: K }>["value"];

const WS_BC_NAME = "ws-control";
const WS_RECONNECT_DELAY = 2_000;
const WS_MAX_RECONNECT_DELAY = 30_000;
const WS_MAX_RECONNECT_ATTEMPTS = 5;
// サーバーの停止（1001 Going Away）では他のレプリカへすぐつなぎ直す。一斉に来ないよう散らす
const WS_GOING_AWAY_DELAY_MAX = 1_000;
// サーバーが閉じるときのコード。4401 は認証の失効、4403 はワークスペースから外されたとき
const WS_CLOSE_UNAUTHENTICATED = 4401;
const WS_CLOSE_FORBIDDEN = 4403;

const getWsUrl = (ticket: string) =>
  `${import.meta.env.VITE_WS_URL}/ws?ticket=${encodeURIComponent(ticket)}`;

const parseServerEvent = (data: string) => {
  try {
    return fromJsonString(ServerEventSchema, data, { ignoreUnknownFields: true });
  } catch (error) {
    console.warn("WebSocketイベントの形式が想定と異なります:", error);
    return null;
  }
};

export class WsClient {
  private ws: WebSocket | null = null;
  private reconnectTimeoutId: ReturnType<typeof setTimeout> | null = null;
  private reconnectAttempts = 0;
  // チケットの取得を待つ間に切断や再接続が起きたら、古い接続の続きを捨てる
  private connectionId = 0;
  private isActiveLeader = false;
  private isClosed = false;
  private readonly bc = new BroadcastChannel(WS_BC_NAME);
  // 同じチャンネルを複数の画面が購読するため参照数で持ち、再接続時に送り直す
  private readonly joinedChannels = new Map<string, number>();
  private viewingChannelId = "";
  // 2 回目以降の接続で、切断中に届かなかったイベントを取り直させる
  private hasOpened = false;
  private readonly reconnectHandlers = new Set<() => void>();

  // 型を case ごとに対応づけるため、Map ではなく全 case を持つオブジェクトにする
  private readonly handlers: { [K in WsEventType]: Set<(payload: WsEventPayload<K>) => void> } = {
    channelViewers: new Set(),
    customEmojiCreated: new Set(),
    customEmojiDeleted: new Set(),
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

  // 接続のたびに 1 回限りのチケットを発行する
  private readonly issueTicket: () => Promise<string>;
  // 隠れている間も切断しない。プッシュ通知の代わりに OS の通知を出すデスクトップアプリ向け
  private readonly keepAliveWhenHidden: boolean;

  public constructor(issueTicket: () => Promise<string>, keepAliveWhenHidden: boolean) {
    this.issueTicket = issueTicket;
    this.keepAliveWhenHidden = keepAliveWhenHidden;
    // 他タブが接続を始めたらリーダーを譲る
    this.bc.addEventListener("message", () => {
      this.disconnect();
    });
    window.addEventListener("visibilitychange", this.handleVisibility, false);
    window.addEventListener("focus", this.handleFocus, false);
    window.addEventListener("online", this.handleFocus, false);
    window.addEventListener("beforeunload", this.handleUnload, false);
    if (keepAliveWhenHidden || (document.visibilityState === "visible" && document.hasFocus())) {
      this.becomeLeaderAndConnect();
    }
  }

  /** サーバーイベントを購読中のハンドラへ配る（WebSocket の message ハンドラ） */
  public readonly eventDispatcher = (event: MessageEvent<string>) => {
    const oneof = parseServerEvent(event.data)?.event;
    if (oneof?.case === undefined) {
      return;
    }

    this.emit(oneof);
  };

  private emit<T extends WsEventType>(
    event: { [K in T]: { case: K; value: WsEventPayload<K> } }[T],
  ) {
    for (const handler of this.handlers[event.case]) {
      handler(event.value);
    }
  }

  /** サーバーイベントの購読を開始する。戻り値を呼ぶと購読を解除する */
  public on<T extends WsEventType>(type: T, cb: (payload: WsEventPayload<T>) => void) {
    this.handlers[type].add(cb);
    return () => {
      this.handlers[type].delete(cb);
    };
  }

  /** 切断を挟んで再びつながったときに呼ぶ。戻り値を呼ぶと解除する */
  public onReconnect(cb: () => void) {
    this.reconnectHandlers.add(cb);
    return () => {
      this.reconnectHandlers.delete(cb);
    };
  }

  private async connect() {
    const id = ++this.connectionId;
    try {
      const ticket = await this.issueTicket();
      if (id !== this.connectionId || !this.isActiveLeader) {
        return;
      }
      this.ws = new WebSocket(getWsUrl(ticket));
      this.ws.addEventListener("open", this.onOpen);
      this.ws.addEventListener("close", this.onClose);
      this.ws.addEventListener("message", this.eventDispatcher);
    } catch (error) {
      if (id === this.connectionId) {
        console.warn("WebSocketに接続できませんでした", error);
        this.scheduleReconnect(false);
      }
    }
  }

  private readonly onOpen = () => {
    this.reconnectAttempts = 0;
    for (const channelId of this.joinedChannels.keys()) {
      this.send({ case: "joinChannel", value: { channelId } });
    }
    if (this.viewingChannelId !== "") {
      this.send({ case: "viewChannel", value: { channelId: this.viewingChannelId } });
    }
    if (this.hasOpened) {
      for (const handler of this.reconnectHandlers) {
        handler();
      }
    }
    this.hasOpened = true;
  };

  // error のあとには必ず close が来るため、失敗はここでだけ数える
  private readonly onClose = (event: CloseEvent) => {
    this.ws = null;
    if (!this.isActiveLeader || event.code === 1000) {
      return;
    }
    if (event.code === WS_CLOSE_FORBIDDEN) {
      this.close();
      navigateTo({ to: "/app" });
      return;
    }
    if (event.code === WS_CLOSE_UNAUTHENTICATED) {
      refreshOrSignOut().then(
        () => {
          void this.connect();
        },
        () => {
          this.scheduleReconnect(false);
        },
      );
      return;
    }
    this.scheduleReconnect(event.code === 1001);
  };

  /** 指数バックオフ（2, 4, 8, 16 秒、最大 30 秒）でつなぎ直す。上限に達したらリーダーを降り、focus か online で再開する */
  private scheduleReconnect(isGoingAway: boolean) {
    if (!this.isActiveLeader || this.reconnectTimeoutId !== null) {
      return;
    }
    this.reconnectAttempts += 1;
    if (this.reconnectAttempts >= WS_MAX_RECONNECT_ATTEMPTS) {
      console.error("WebSocketの再接続を諦めました", { attempts: this.reconnectAttempts });
      this.disconnect();
      return;
    }
    this.reconnectTimeoutId = setTimeout(
      () => {
        this.reconnectTimeoutId = null;
        void this.connect();
      },
      isGoingAway
        ? Math.random() * WS_GOING_AWAY_DELAY_MAX
        : Math.min(WS_RECONNECT_DELAY * 2 ** (this.reconnectAttempts - 1), WS_MAX_RECONNECT_DELAY),
    );
  }

  public joinChannel(channelId: string) {
    const count = this.joinedChannels.get(channelId) ?? 0;
    this.joinedChannels.set(channelId, count + 1);
    if (count === 0) {
      this.send({ case: "joinChannel", value: { channelId } });
    }
  }
  public leaveChannel(channelId: string) {
    const count = this.joinedChannels.get(channelId) ?? 0;
    if (count > 1) {
      this.joinedChannels.set(channelId, count - 1);
      return;
    }
    this.joinedChannels.delete(channelId);
    this.send({ case: "leaveChannel", value: { channelId } });
  }
  /** 閲覧中のチャンネルを通知する。空文字で閲覧をやめる */
  public viewChannel(channelId: string) {
    this.viewingChannelId = channelId;
    this.send({ case: "viewChannel", value: { channelId } });
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

  /** 接続を閉じる。タブが見えるようになればまたつなぎ直せるよう、イベントの監視は続ける */
  private disconnect() {
    this.isActiveLeader = false;
    this.connectionId += 1;
    if (this.ws) {
      this.ws.removeEventListener("open", this.onOpen);
      this.ws.removeEventListener("close", this.onClose);
      this.ws.removeEventListener("message", this.eventDispatcher);
      this.ws.close();
      this.ws = null;
    }
    if (this.reconnectTimeoutId !== null) {
      clearTimeout(this.reconnectTimeoutId);
      this.reconnectTimeoutId = null;
    }
    this.reconnectAttempts = 0;
  }

  public close() {
    this.isClosed = true;
    this.disconnect();
    window.removeEventListener("visibilitychange", this.handleVisibility, false);
    window.removeEventListener("focus", this.handleFocus, false);
    window.removeEventListener("online", this.handleFocus, false);
    window.removeEventListener("beforeunload", this.handleUnload, false);
    this.bc.close();
  }

  private readonly handleVisibility = () => {
    if (document.visibilityState === "visible") {
      this.becomeLeaderAndConnect();
    } else if (!this.keepAliveWhenHidden) {
      this.disconnect();
    }
  };

  private readonly handleFocus = () => {
    this.becomeLeaderAndConnect();
  };

  private readonly handleUnload = () => {
    this.close();
  };

  private becomeLeaderAndConnect() {
    if (this.isActiveLeader || this.isClosed) {
      return;
    }
    this.isActiveLeader = true;
    this.bc.postMessage({ type: "ws_active" });
    void this.connect();
  }
}
