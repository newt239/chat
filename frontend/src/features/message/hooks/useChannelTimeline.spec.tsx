import type { ReactNode } from "react";

import { create, toJsonString } from "@bufbuild/protobuf";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { ServerEventSchema } from "#/gen/chat/v1/event_pb";
import { ListMessagesResponseSchema } from "#/gen/chat/v1/message_service_pb";
import { WsClient } from "#/lib/ws";
import { WsClientContext } from "#/providers/ws/useWsClient";

import { useChannelTimeline } from "./useChannelTimeline";
import { messagePagesKey } from "./useMessagePages";

import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";

import type { MessageInitShape } from "@bufbuild/protobuf";
import type { InfiniteData } from "@tanstack/react-query";

const dispatch = (client: WsClient, init: MessageInitShape<typeof ServerEventSchema>) => {
  client.eventDispatcher(
    new MessageEvent("message", {
      data: toJsonString(ServerEventSchema, create(ServerEventSchema, init)),
    }),
  );
};

describe("useChannelTimeline", () => {
  test("返信が届くと親の返信数を進め、編集されても返信数を保つ", () => {
    const client = new WsClient(() => Promise.resolve("ticket"), false);
    const queryClient = new QueryClient();
    const key = messagePagesKey("ch1", false);
    queryClient.setQueryData(key, {
      pageParams: [null],
      pages: [
        create(ListMessagesResponseSchema, {
          messages: [{ content: { case: "userMessage", value: { channelId: "ch1", id: "p1" } } }],
        }),
      ],
    });
    const wrapper = ({ children }: { children: ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        <WsClientContext value={client}>{children}</WsClientContext>
      </QueryClientProvider>
    );
    renderHook(
      () =>
        useChannelTimeline({
          channelId: "ch1",
          descendantIds: [],
          includeDescendants: false,
          items: [],
        }),
      { wrapper },
    );
    const parent = () => {
      const item =
        queryClient.getQueryData<InfiniteData<ListMessagesResponse>>(key)?.pages[0]?.messages[0];
      return item?.content.case === "userMessage" ? item.content.value : undefined;
    };

    act(() => {
      for (const id of ["r1", "r2"]) {
        dispatch(client, {
          event: {
            case: "newMessage",
            value: {
              channelId: "ch1",
              message: { channelId: "ch1", id, parentId: "p1", userId: "u2" },
            },
          },
        });
      }
    });
    expect(parent()?.threadMetadata?.replyCount).toBe(2);

    act(() => {
      dispatch(client, {
        event: {
          case: "messageUpdated",
          value: { channelId: "ch1", message: { body: "編集", channelId: "ch1", id: "p1" } },
        },
      });
    });
    expect(parent()?.body).toBe("編集");
    expect(parent()?.threadMetadata?.replyCount).toBe(2);
    client.close();
  });
});
