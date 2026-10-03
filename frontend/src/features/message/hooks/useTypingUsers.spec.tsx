import type { ReactNode } from "react";

import { create, toJsonString } from "@bufbuild/protobuf";
import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, test, vi } from "vite-plus/test";

import { ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { WsClient } from "#/lib/ws";
import { WsClientContext } from "#/providers/ws/useWsClient";

import { useTypingUsers } from "./useTypingUsers";

const typingEvent = (kind: "typing" | "stopTyping", channelId: string, userId: string) =>
  new MessageEvent("message", {
    data: toJsonString(
      ServerEventSchema,
      create(ServerEventSchema, { event: { case: kind, value: { channelId, userId } } }),
    ),
  });

const renderTypingUsers = () => {
  const client = new WsClient(() => Promise.resolve("ticket"), false);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <WsClientContext value={client}>{children}</WsClientContext>
  );
  const { result } = renderHook(() => useTypingUsers("ch1"), { wrapper });
  return { client, result };
};

afterEach(() => {
  vi.useRealTimers();
});

describe("useTypingUsers", () => {
  test("表示中のチャンネルで入力中のユーザーだけを返す", () => {
    const { client, result } = renderTypingUsers();

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

  test("入力中の通知が途絶えたら終了の通知がなくても外す", () => {
    vi.useFakeTimers();
    const { client, result } = renderTypingUsers();

    act(() => {
      client.eventDispatcher(typingEvent("typing", "ch1", "u1"));
      vi.advanceTimersByTime(4_000);
      client.eventDispatcher(typingEvent("typing", "ch1", "u1"));
      client.eventDispatcher(typingEvent("typing", "ch1", "u2"));
      vi.advanceTimersByTime(4_000);
    });
    expect(result.current).toEqual(["u1", "u2"]);

    act(() => {
      vi.advanceTimersByTime(2_000);
    });
    expect(result.current).toEqual([]);
    client.close();
  });
});
