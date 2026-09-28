import { useEffect, useState } from "react";

import { IconMessages } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { EmptyState } from "#/components/ui/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { ThreadCard } from "#/features/thread/components/ThreadCard";
import { useParticipatingThreads } from "#/features/thread/hooks/useParticipatingThreads";

import type { Message } from "#/gen/chat/v1/message_pb";
import type { ParticipatingThread, ThreadCursor } from "#/gen/chat/v1/thread_service_pb";

type ThreadItem = ParticipatingThread & { firstMessage: Message };

export const ThreadListPage = () => {
  const { t } = useTranslation();
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

  const handleMarkedRead = (threadId: string) => {
    setItems((prev) =>
      prev.map((item) => (item.threadId === threadId ? { ...item, unreadCount: 0 } : item)),
    );
    void refetch();
  };

  return (
    <>
      <PageHeader icon={<IconMessages />} title={t("shell.nav.threads")} />
      <div className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto p-4">
        {isLoading && items.length === 0 ? (
          <Skeleton className="h-20 w-full" />
        ) : items.length === 0 ? (
          <EmptyState
            icon={<IconMessages />}
            title={t("shell.thread.emptyTitle")}
            description={t("shell.thread.emptyDescription")}
          />
        ) : (
          items.map((item) => (
            <ThreadCard
              key={item.threadId}
              workspaceId={workspaceId}
              thread={item}
              onMarkedRead={handleMarkedRead}
            />
          ))
        )}
        {next && (
          <Button
            variant="secondary"
            className="self-center"
            isPending={isFetching}
            onPress={() => {
              setCursor(next);
            }}
          >
            {t("shell.thread.loadMore")}
          </Button>
        )}
      </div>
    </>
  );
};
