import type { ReactNode } from "react";

import { create, toJsonString } from "@bufbuild/protobuf";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { WsClient } from "#/lib/ws";
import { WsClientContext } from "#/providers/ws/wsClientContext";

import { useTypingUsers } from "./useTypingUsers";

const typingEvent = (kind: "typing" | "stopTyping", channelId: string, userId: string) =>
  new MessageEvent("message", {
    data: toJsonString(
      ServerEventSchema,
      create(ServerEventSchema, { event: { case: kind, value: { channelId, userId } } }),
    ),
  });

describe("useTypingUsers", () => {
  test("表示中のチャンネルで入力中のユーザーだけを返す", () => {
    const client = new WsClient("token", "ws1");
    const wrapper = ({ children }: { children: ReactNode }) => (
      <WsClientContext value={{ wsClient: client }}>{children}</WsClientContext>
    );
    const { result } = renderHook(() => useTypingUsers("ch1"), { wrapper });

    act(() => {
      client.eventDispatcher(typingEvent("typing", "ch1", "u1"));
      client.eventDispatcher(typingEvent("typing", "ch1", "u1"));
      client.eventDispatcher(typingEvent("typing", "ch2", "u2"));
      client.eventDispatcher(typingEvent("typing", "ch1", "u3"));
    });
    expect(result.current).toEqual(["u1", "u3"]);

    act(() => {
      client.eventDispatcher(typingEvent("stopTyping", "ch1", "u1"));
    });
    expect(result.current).toEqual(["u3"]);
    client.close();
  });
});
