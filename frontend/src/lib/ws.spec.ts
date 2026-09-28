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

const setupSocket = () => {
  const sent: string[] = [];
  const openListeners: (() => void)[] = [];
  vi.stubGlobal(
    "WebSocket",
    class {
      public static readonly OPEN = 1;
      public readonly readyState = 1;
      public addEventListener(type: string, listener: () => void) {
        if (type === "open") {
          openListeners.push(listener);
        }
      }
      public removeEventListener() {}
      public send(data: string) {
        sent.push(data);
      }
      public close() {}
    },
  );
  const client = new WsClient("token", "ws1");
  globalThis.dispatchEvent(new Event("focus"));
  return {
    client,
    open: () => {
      for (const listener of openListeners) {
        listener();
      }
    },
    sent,
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
