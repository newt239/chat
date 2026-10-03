import type { ReactNode } from "react";

import { create } from "@bufbuild/protobuf";
import { createRouterTransport } from "@connectrpc/connect";
import { createConnectQueryKey, TransportProvider } from "@connectrpc/connect-query";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { toast } from "#/components/ui/ToastRegion/toast";
import { ListMessagesResponseSchema, MessageService } from "#/gen/chat/v1/message_service_pb";
import {
  GetThreadMetadataResponseSchema,
  ListParticipatingThreadsResponseSchema,
  ThreadService,
} from "#/gen/chat/v1/thread_service_pb";

import { useToggleThreadFollow } from "./useToggleThreadFollow";

import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";
import type {
  GetThreadMetadataResponse,
  ListParticipatingThreadsResponse,
} from "#/gen/chat/v1/thread_service_pb";

import type { InfiniteData } from "@tanstack/react-query";

vi.mock("#/components/ui/ToastRegion/toast", () => ({ toast: vi.fn() }));

const metadataKey = createConnectQueryKey({
  cardinality: "finite",
  input: { messageId: "t1" },
  schema: ThreadService.method.getThreadMetadata,
});
const timelineKey = createConnectQueryKey({
  cardinality: "infinite",
  input: { channelId: "c1" },
  schema: MessageService.method.listMessages,
});
const threadListKey = createConnectQueryKey({
  cardinality: "infinite",
  input: { workspaceId: "ws1" },
  schema: ThreadService.method.listParticipatingThreads,
});

const userMessage = (id: string) => ({
  content: {
    case: "userMessage" as const,
    value: { id, threadMetadata: { messageId: id, replyCount: 1 } },
  },
});

const setup = () => {
  const follow = vi.fn(() => ({}));
  const queryClient = new QueryClient();
  queryClient.setQueryData(
    metadataKey,
    create(GetThreadMetadataResponseSchema, { metadata: { messageId: "t1", replyCount: 2 } }),
  );
  queryClient.setQueryData(timelineKey, {
    pageParams: [null],
    pages: [
      create(ListMessagesResponseSchema, { messages: [userMessage("t1"), userMessage("m2")] }),
    ],
  });
  queryClient.setQueryData(threadListKey, {
    pageParams: [undefined],
    pages: [create(ListParticipatingThreadsResponseSchema, { threads: [{ threadId: "t1" }] })],
  });
  const transport = createRouterTransport((router) => {
    router.rpc(ThreadService.method.followThread, follow);
  });
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <TransportProvider transport={transport}>{children}</TransportProvider>
    </QueryClientProvider>
  );
  const { result } = renderHook(() => useToggleThreadFollow("t1"), { wrapper });
  return { follow, queryClient, result };
};

describe("useToggleThreadFollow", () => {
  test("フォローするとメタデータ・タイムライン・参加中の一覧を更新して通知する", async () => {
    const { follow, queryClient, result } = setup();

    act(() => {
      result.current.setFollowing(true);
    });

    await waitFor(() => {
      expect(toast).toHaveBeenCalledWith("スレッドをフォローしました");
    });
    expect(follow).toHaveBeenCalledOnce();
    expect(
      queryClient.getQueryData<GetThreadMetadataResponse>(metadataKey)?.metadata?.isFollowing,
    ).toBe(true);
    expect(
      queryClient
        .getQueryData<InfiniteData<ListMessagesResponse>>(timelineKey)
        ?.pages[0]?.messages.map((item) =>
          item.content.case === "userMessage"
            ? item.content.value.threadMetadata?.isFollowing
            : null,
        ),
    ).toEqual([true, false]);
    expect(
      queryClient.getQueryData<InfiniteData<ListParticipatingThreadsResponse>>(threadListKey)
        ?.pages[0]?.threads[0]?.isFollowing,
    ).toBe(true);
  });
});
