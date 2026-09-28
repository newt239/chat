import { useEffect, useMemo, useState } from "react";

import { Button, Loader, Stack, Text } from "@mantine/core";

import { ThreadCard } from "#/features/thread/components/ThreadCard";
import { useParticipatingThreads } from "#/features/thread/hooks/useParticipatingThreads";
import { useWorkspaceId } from "#/lib/routeParams";

import type { ParticipatingThread } from "#/features/thread/schemas";

export const ThreadListPage = () => {
  const workspaceId = useWorkspaceId();

  const [cursorLastActivityAt, setCursorLastActivityAt] = useState<string | undefined>();
  const [cursorThreadId, setCursorThreadId] = useState<string | undefined>();

  const { data, isLoading, isFetching, refetch } = useParticipatingThreads({
    cursorLastActivityAt,
    cursorThreadId,
    limit: 20,
    workspaceId,
  });

  // ページを跨いで結果を積み上げる
  const [items, setItems] = useState<ParticipatingThread[]>([]);

  useEffect(() => {
    if (!data) {
      return;
    }
    setItems((prev) => {
      const known = new Set(prev.map((item) => item.thread_id));
      return [...prev, ...data.items.filter((item) => !known.has(item.thread_id))];
    });
  }, [data]);

  useEffect(() => {
    setItems([]);
  }, [workspaceId]);

  const next = data?.next_cursor;

  const isBusy = isLoading || isFetching;

  const handleLoadMore = () => {
    if (!next) {
      return;
    }
    setCursorLastActivityAt(next.last_activity_at);
    setCursorThreadId(next.thread_id);
  };

  const handleMarkedRead = (threadId: string) => {
    setItems((prev) =>
      prev.map((item) => (item.thread_id === threadId ? { ...item, unread_count: 0 } : item)),
    );
    void refetch();
  };

  const empty = useMemo(() => !isBusy && items.length === 0, [isBusy, items.length]);

  return (
    <div className="p-3">
      <Stack gap={12}>
        <Text fw={700} size="lg">
          参加中のスレッド
        </Text>
        {isBusy && items.length === 0 ? (
          <div className="flex justify-center py-8">
            <Loader />
          </div>
        ) : empty ? (
          <div className="text-center text-gray-600 py-10">参加中のスレッドはありません</div>
        ) : (
          <Stack gap={8}>
            {items.map((t) => (
              <ThreadCard key={t.thread_id} thread={t} onMarkedRead={handleMarkedRead} />
            ))}
          </Stack>
        )}

        <div className="flex justify-center py-2">
          <Button onClick={handleLoadMore} disabled={!next || isBusy} variant="light">
            さらに読み込む
          </Button>
        </div>
      </Stack>
    </div>
  );
};
