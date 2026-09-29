import { IconMessages } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { ThreadCard } from "#/features/thread/components/ThreadCard";
import { useParticipatingThreads } from "#/features/thread/hooks/useParticipatingThreads";
import { useLoadMoreRef } from "#/hooks/useLoadMoreRef";

export const ThreadListPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const {
    data: threads,
    isLoading,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = useParticipatingThreads(workspaceId);
  const loadMoreRef = useLoadMoreRef(() => {
    void fetchNextPage();
  }, hasNextPage && !isFetchingNextPage);

  return (
    <>
      <PageHeader icon={<IconMessages />} title={t("shell.nav.threads")}>
        <span className="hidden text-caption text-muted md:inline">{t("inbox.replyHint")}</span>
      </PageHeader>
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-[18px] py-3 max-md:px-2.5">
        {isLoading ? (
          <Skeleton className="h-32 w-full rounded-[10px]" />
        ) : threads === undefined || threads.length === 0 ? (
          <EmptyState
            icon={<IconMessages />}
            title={t("shell.thread.emptyTitle")}
            description={t("shell.thread.emptyDescription")}
          />
        ) : (
          threads.map((thread) => (
            <ThreadCard key={thread.threadId} workspaceId={workspaceId} thread={thread} />
          ))
        )}
        <div ref={loadMoreRef}>
          {isFetchingNextPage && <Skeleton className="h-32 w-full rounded-[10px]" />}
        </div>
      </div>
    </>
  );
};
