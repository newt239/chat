import { useInfiniteQuery } from "@tanstack/react-query";

import type { Timestamp } from "@bufbuild/protobuf/wkt";
import type { QueryKey } from "@tanstack/react-query";

export type Direction = "older" | "newer";

// 最初のページは null。前後のページは端の項目の日時を境に取る
export type PageCursor = { direction: Direction; boundary: Timestamp } | null;

/** 前後のページのカーソルをリクエストの until / since に変える */
export const toRange = ({ direction, boundary }: NonNullable<PageCursor>) =>
  direction === "older" ? { until: boundary } : { since: boundary };

type Page = { hasMore: boolean; hasNewer: boolean };

const hasMoreToward = (page: Page, direction: Direction) =>
  direction === "older" ? page.hasMore : page.hasNewer;

const cursorAt = (direction: Direction, item: { createdAt?: Timestamp } | undefined) =>
  item?.createdAt === undefined ? undefined : { boundary: item.createdAt, direction };

type UseBidirectionalPagesArgs<P extends Page, Item extends { createdAt?: Timestamp }> = {
  queryKey: QueryKey;
  enabled: boolean;
  fetchPage: (cursor: PageCursor, signal: AbortSignal) => Promise<P>;
  // 並びの順で返す
  getItems: (page: P) => Item[];
  // 項目が新しい順に並ぶ（チャンネル）か、古い順に並ぶ（スレッド）か
  newestFirst: boolean;
};

/** 最初に取得した範囲の前後を、スクロールに合わせてキャッシュのページとして足していく */
export const useBidirectionalPages = <P extends Page, Item extends { createdAt?: Timestamp }>({
  queryKey,
  enabled,
  fetchPage,
  getItems,
  newestFirst,
}: UseBidirectionalPagesArgs<P, Item>) => {
  // 配列の末尾に足す方向と先頭に足す方向
  const endward: Direction = newestFirst ? "older" : "newer";
  const startward: Direction = newestFirst ? "newer" : "older";

  const query = useInfiniteQuery({
    enabled,
    getNextPageParam: (last: P) =>
      hasMoreToward(last, endward) ? cursorAt(endward, getItems(last).at(-1)) : undefined,
    getPreviousPageParam: (first: P) =>
      hasMoreToward(first, startward) ? cursorAt(startward, getItems(first).at(0)) : undefined,
    initialPageParam: null,
    queryFn: ({ pageParam, signal }: { pageParam: PageCursor; signal: AbortSignal }) =>
      fetchPage(pageParam, signal),
    queryKey,
  });

  const pages = query.data?.pages;
  const items = pages?.flatMap((page) => getItems(page));

  // スクロールのたびに呼ばれるため、読み込み中は重ねて取得しない
  const load = (direction: Direction) => {
    if (query.isFetching) {
      return;
    }
    void (direction === endward ? query.fetchNextPage() : query.fetchPreviousPage());
  };

  const isFetchingEnd = query.isFetchingNextPage;
  const isFetchingStart = query.isFetchingPreviousPage;
  const hasEnd = query.hasNextPage;
  const hasStart = query.hasPreviousPage;

  return {
    error: query.error,
    hasNewer: newestFirst ? hasStart : hasEnd,
    hasOlder: newestFirst ? hasEnd : hasStart,
    isError: query.isError,
    isLoading: query.isLoading,
    items,
    load,
    loading: isFetchingEnd ? endward : isFetchingStart ? startward : null,
    pages,
  };
};
