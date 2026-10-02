import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import { callUnaryMethod, createConnectQueryKey, useTransport } from "@connectrpc/connect-query";

import { MessageService } from "#/gen/chat/v1/message_service_pb";

import { toRange, useBidirectionalPages } from "./useBidirectionalPages";

import type { PageCursor } from "./useBidirectionalPages";

import type { ListMessagesResponse } from "#/gen/chat/v1/message_service_pb";

const MESSAGES_PAGE_SIZE = 50;

const getMessages = (page: ListMessagesResponse) => page.messages;

/** WebSocket の差分をこのチャンネルのタイムラインのキャッシュに当てるためのキー */
export const messagePagesKey = (channelId: string, includeDescendants: boolean) =>
  createConnectQueryKey({
    cardinality: "infinite",
    input: { channelId, includeDescendants },
    schema: MessageService.method.listMessages,
  });

type UseMessagePagesArgs = {
  // チャンネルの解決を待つ間は null
  channelId: string | null;
  includeDescendants: boolean;
  // 渡すとその日時の前後から読む
  around: Date | null;
};

/** チャンネルのメッセージを新しい順に取得し、スクロールに合わせて前後を足す */
export const useMessagePages = ({ channelId, includeDescendants, around }: UseMessagePagesArgs) => {
  const transport = useTransport();
  const input = {
    around: around === null ? undefined : timestampFromDate(around),
    channelId: channelId ?? "",
    includeDescendants,
    limit: MESSAGES_PAGE_SIZE,
  };

  return useBidirectionalPages({
    enabled: channelId !== null,
    fetchPage: (cursor: PageCursor, signal) =>
      callUnaryMethod(
        transport,
        MessageService.method.listMessages,
        cursor === null ? input : { ...input, around: undefined, ...toRange(cursor) },
        { signal },
      ),
    getItems: getMessages,
    newestFirst: true,
    queryKey: createConnectQueryKey({
      cardinality: "infinite",
      input,
      schema: MessageService.method.listMessages,
      transport,
    }),
  });
};
