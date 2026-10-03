import { IconAt } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { useMentions } from "#/features/inbox/hooks/useMentions";
import { useLoadMoreRef } from "#/hooks/useLoadMoreRef";

import { MentionCard } from "./MentionCard";

export const MentionsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const mentions = useMentions(workspaceId);
  const { data: messages, isLoading, isError, isFetchingNextPage } = mentions;
  const loadMoreRef = useLoadMoreRef(mentions);

  return (
    <>
      <PageHeader icon={<IconAt />} title={t("shell.nav.mentions")} />
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-4.5 py-3 max-md:px-2.5">
        {isLoading ? (
          <Skeleton className="h-32 w-full rounded-lg" />
        ) : isError ? (
          <p role="alert" className="m-0 p-4 text-caption text-danger">
            {t("inbox.mention.failed")}
          </p>
        ) : messages === undefined || messages.length === 0 ? (
          <EmptyState
            icon={<IconAt />}
            title={t("inbox.mention.emptyTitle")}
            description={t("inbox.mention.emptyDescription")}
          />
        ) : (
          messages.map((message) => (
            <MentionCard key={message.id} workspaceId={workspaceId} message={message} />
          ))
        )}
        <div ref={loadMoreRef}>
          {isFetchingNextPage && <Skeleton className="h-32 w-full rounded-lg" />}
        </div>
      </div>
    </>
  );
};
