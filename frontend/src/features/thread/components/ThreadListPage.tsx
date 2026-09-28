import { useEffect, useMemo, useState } from "react";

import { Button, Loader, Stack, Text } from "@mantine/core";
import { useParams } from "@tanstack/react-router";

import { ThreadCard } from "#/features/thread/components/ThreadCard";
import { useParticipatingThreads } from "#/features/thread/hooks/useParticipatingThreads";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread, ThreadCursor } from "#/gen/chat/v1/thread_service_pb";

type ThreadItem = ParticipatingThread & { firstMessage: Message };

export const ThreadListPage = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });

  const [cursor, setCursor] = useState<ThreadCursor>();

  const { data, isLoading, isFetching, refetch } = useParticipatingThreads(workspaceId, cursor);

  // ページを跨いで結果を積み上げる
  const [items, setItems] = useState<ThreadItem[]>([]);

  useEffect(() => {
    if (!data) {
      return;
    }
    setItems((prev) => {
      const known = new Set(prev.map((item) => item.threadId));
      return [...prev, ...data.threads.filter((item) => !known.has(item.threadId))];
    });
  }, [data]);

  useEffect(() => {
    setItems([]);
  }, [workspaceId]);

  const next = data?.nextCursor;

  const isBusy = isLoading || isFetching;

  const handleLoadMore = () => {
    if (!next) {
      return;
    }
    setCursor(next);
  };

  const handleMarkedRead = (threadId: string) => {
    setItems((prev) =>
      prev.map((item) => (item.threadId === threadId ? { ...item, unreadCount: 0 } : item)),
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
              <ThreadCard key={t.threadId} thread={t} onMarkedRead={handleMarkedRead} />
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
