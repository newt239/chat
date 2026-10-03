import { IconBookmark } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { MessageLinkCard } from "#/features/message/components/MessageLinkCard";

import { useBookmarks } from "../hooks/useBookmarks";

export const BookmarksPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: bookmarks, isLoading, error } = useBookmarks();

  const renderBody = () => {
    if (isLoading) {
      return (
        <div className="flex flex-col gap-2 p-3">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-12 w-full" />
        </div>
      );
    }
    if (error) {
      return <p className="m-0 p-4 font-sans text-body text-danger">{t("bookmark.loadFailed")}</p>;
    }
    if (!bookmarks || bookmarks.length === 0) {
      return (
        <EmptyState
          icon={<IconBookmark />}
          title={t("bookmark.empty")}
          description={t("bookmark.emptyHint")}
        />
      );
    }
    return (
      <div className="flex flex-col gap-0.5 p-1.5">
        {bookmarks.map((bookmark) => (
          <MessageLinkCard
            key={bookmark.message.id}
            message={bookmark.message}
            workspaceId={workspaceId}
            markedAt={bookmark.createdAt}
          />
        ))}
      </div>
    );
  };

  return (
    <>
      <PageHeader icon={<IconBookmark />} title={t("shell.nav.bookmarks")} />
      <div className="flex min-h-0 flex-1 flex-col overflow-y-auto">{renderBody()}</div>
    </>
  );
};
