import { describe, expect, test } from "vite-plus/test";

import { WsClient } from "#/lib/ws";

const serverEvent = (type: string, payload: unknown) =>
  new MessageEvent("message", { data: JSON.stringify({ payload, type }) });

const newMessagePayload = {
  channel_id: "ch1",
  message: {
    body: "hello",
    channelId: "ch1",
    createdAt: "2026-01-01T00:00:00Z",
    id: "m1",
    isDeleted: false,
    user: { displayName: "Alice", id: "u1" },
    userId: "u1",
  },
};

describe("WsClient のイベント購読", () => {
  test("購読したイベントだけにペイロードが届く", () => {
    const client = new WsClient("token", "ws1");
    const received: string[] = [];
    const other: string[] = [];

    client.on("new_message", (payload) => {
      received.push(payload.message.id);
    });
    client.on("message_deleted", (payload) => {
      other.push(payload.deleteData.id);
    });

    client.eventDispatcher(serverEvent("new_message", newMessagePayload));

    expect(received).toEqual(["m1"]);
    expect(other).toEqual([]);
    client.close();
  });

  test("戻り値を呼ぶと購読が解除される", () => {
    const client = new WsClient("token", "ws1");
    const received: string[] = [];

    const unsubscribe = client.on("new_message", (payload) => {
      received.push(payload.message.id);
    });

    client.eventDispatcher(serverEvent("new_message", newMessagePayload));
    unsubscribe();
    client.eventDispatcher(serverEvent("new_message", newMessagePayload));

    expect(received).toEqual(["m1"]);
    client.close();
  });

  test("形式が不正なイベントはハンドラを呼ばない", () => {
    const client = new WsClient("token", "ws1");
    let called = false;

    client.on("typing", () => {
      called = true;
    });

    client.eventDispatcher(serverEvent("typing", { channel_id: "ch1" }));

    expect(called).toBe(false);
    client.close();
  });
});
