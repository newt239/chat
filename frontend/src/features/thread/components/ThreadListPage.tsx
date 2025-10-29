import { useMemo, useState } from "react";

import { Button, Loader, Stack, Text } from "@mantine/core";
import { useParams } from "@tanstack/react-router";

import { ThreadCard } from "@/features/thread/components/ThreadCard";
import { useParticipatingThreads } from "@/features/thread/hooks/useParticipatingThreads";

export const ThreadListPage = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId/threads" });

  const [cursorLastActivityAt, setCursorLastActivityAt] = useState<string | undefined>(undefined);
  const [cursorThreadId, setCursorThreadId] = useState<string | undefined>(undefined);

  const { data, isLoading, isFetching, refetch } = useParticipatingThreads({
    workspaceId,
    cursorLastActivityAt,
    cursorThreadId,
    limit: 20,
  });

  const items = data?.items ?? [];
  const next = data?.next_cursor;

  const isBusy = isLoading || isFetching;

  const handleLoadMore = () => {
    if (!next) return;
    setCursorLastActivityAt(next.last_activity_at);
    setCursorThreadId(next.thread_id);
    // 直後のuseQueryはキーが変わるため自動再取得される
  };

  const handleMarkedRead = () => {
    // 楽観的に未読数を0にしたUIにするにはローカル状態で持ち替えるが、まずはrefetchで簡易更新
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
              <ThreadCard key={`${t.thread_id}`} thread={t} onMarkedRead={handleMarkedRead} />
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
