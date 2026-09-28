import { create, toJsonString } from "@bufbuild/protobuf";
import { describe, expect, test } from "vite-plus/test";

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
