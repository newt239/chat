import { create, toJsonString } from "@bufbuild/protobuf";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { WsClient } from "#/lib/ws";

import type { MessageInitShape } from "@bufbuild/protobuf";

const serverEvent = (event: MessageInitShape<typeof ServerEventSchema>["event"]) =>
  new MessageEvent("message", {
    data: toJsonString(ServerEventSchema, create(ServerEventSchema, { event })),
  });

const newMessageEvent = serverEvent({
  case: "newMessage",
  value: { channelId: "ch1", message: { body: "hello", channelId: "ch1", id: "m1" } },
});

describe("WsClient のイベント購読", () => {
  test("購読したイベントだけにペイロードが届く", () => {
    const client = new WsClient("token", "ws1");
    const received: string[] = [];
    const other: string[] = [];

    client.on("newMessage", (payload) => {
      received.push(payload.message?.id ?? "");
    });
    client.on("messageDeleted", (payload) => {
      other.push(payload.messageId);
    });

    client.eventDispatcher(newMessageEvent);

    expect(received).toEqual(["m1"]);
    expect(other).toEqual([]);
    client.close();
  });

  test("戻り値を呼ぶと購読が解除される", () => {
    const client = new WsClient("token", "ws1");
    const received: string[] = [];

    const unsubscribe = client.on("newMessage", (payload) => {
      received.push(payload.message?.id ?? "");
    });

    client.eventDispatcher(newMessageEvent);
    unsubscribe();
    client.eventDispatcher(newMessageEvent);

    expect(received).toEqual(["m1"]);
    client.close();
  });

  test("形式が不正なイベントはハンドラを呼ばない", () => {
    const client = new WsClient("token", "ws1");
    let called = false;

    client.on("typing", () => {
      called = true;
    });

    client.eventDispatcher(new MessageEvent("message", { data: '{"typing":"not-an-object"}' }));

    expect(called).toBe(false);
    client.close();
  });
});

type Listener = (event: CloseEvent) => void;

/** 生成数・送信内容・イベントリスナーを記録する WebSocket に差し替える */
const stubWebSocket = () => {
  const state = { created: 0, listeners: new Map<string, Listener[]>(), sent: [] as string[] };
  vi.stubGlobal(
    "WebSocket",
    class {
      public static readonly OPEN = 1;
      public readonly readyState = 1;
      public constructor() {
        state.created += 1;
      }
      public addEventListener(type: string, listener: Listener) {
        state.listeners.set(type, [...(state.listeners.get(type) ?? []), listener]);
      }
      public removeEventListener() {}
      public send(data: string) {
        state.sent.push(data);
      }
      public close() {}
    },
  );
  const fire = (type: string, event: CloseEvent) => {
    for (const listener of state.listeners.get(type) ?? []) {
      listener(event);
    }
  };
  return { fire, state };
};

const setupSocket = () => {
  const { fire, state } = stubWebSocket();
  const client = new WsClient("token", "ws1");
  globalThis.dispatchEvent(new Event("focus"));
  return {
    client,
    open: () => {
      fire("open", new CloseEvent("open"));
    },
    sent: state.sent,
  };
};

describe("WsClient の購読と閲覧の送信", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("同じチャンネルは最後の購読解除でだけ leaveChannel を送る", () => {
    const { client, sent } = setupSocket();
    client.joinChannel("ch1");
    client.joinChannel("ch1");
    client.leaveChannel("ch1");
    expect(sent).toEqual(['{"joinChannel":{"channelId":"ch1"}}']);
    client.leaveChannel("ch1");
    expect(sent.at(-1)).toBe('{"leaveChannel":{"channelId":"ch1"}}');
    client.close();
  });

  test("接続し直すと購読中のチャンネルと閲覧中のチャンネルを送り直す", () => {
    const { client, open, sent } = setupSocket();
    client.joinChannel("ch1");
    client.viewChannel("ch1");
    sent.length = 0;
    open();
    expect(sent).toEqual([
      '{"joinChannel":{"channelId":"ch1"}}',
      '{"viewChannel":{"channelId":"ch1"}}',
    ]);
    client.close();
  });
});

describe("WsClient の再接続", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.useRealTimers();
  });

  test("最初の接続では呼ばず、つなぎ直したときだけ onReconnect を呼ぶ", () => {
    const { client, open } = setupSocket();
    let reconnected = 0;
    client.onReconnect(() => {
      reconnected += 1;
    });
    open();
    expect(reconnected).toBe(0);
    open();
    expect(reconnected).toBe(1);
    client.close();
  });

  test("タブを隠して戻すとつなぎ直す", () => {
    const { state } = stubWebSocket();
    const client = new WsClient("token", "ws1");
    globalThis.dispatchEvent(new Event("focus"));
    expect(state.created).toBe(1);

    vi.spyOn(document, "visibilityState", "get").mockReturnValue("hidden");
    globalThis.dispatchEvent(new Event("visibilitychange"));
    vi.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
    globalThis.dispatchEvent(new Event("visibilitychange"));
    expect(state.created).toBe(2);
    client.close();
  });

  test("サーバーの停止で閉じられたら 1 秒以内につなぎ直す", () => {
    vi.useFakeTimers();
    const { fire, state } = stubWebSocket();
    const client = new WsClient("token", "ws1");
    globalThis.dispatchEvent(new Event("focus"));
    fire("close", new CloseEvent("close", { code: 1001 }));
    vi.advanceTimersByTime(1_000);
    expect(state.created).toBe(2);
    client.close();
  });
});
