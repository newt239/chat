import { IconAt } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { useMentions } from "#/features/mention/hooks/useMentions";
import { useLoadMoreRef } from "#/hooks/useLoadMoreRef";

import { MentionCard } from "./MentionCard";

type MentionListProps = {
  workspaceId: string;
};

// メンション一覧の中身。デスクトップのメンション画面とモバイルの通知タブで使う
export const MentionList = ({ workspaceId }: MentionListProps) => {
  const { t } = useTranslation();
  const {
    data: messages,
    isLoading,
    isError,
    hasNextPage,
    isFetchingNextPage,
    fetchNextPage,
  } = useMentions(workspaceId);
  const loadMoreRef = useLoadMoreRef(() => {
    void fetchNextPage();
  }, hasNextPage && !isFetchingNextPage);

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-[18px] py-3 max-md:px-2.5">
      {isLoading ? (
        <Skeleton className="h-32 w-full rounded-[10px]" />
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
        {isFetchingNextPage && <Skeleton className="h-32 w-full rounded-[10px]" />}
      </div>
    </div>
  );
};
