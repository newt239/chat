import { IconBookmark } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { MessageLinkCard } from "#/features/message/components/MessageLinkCard";

import { useBookmarks } from "../hooks/useBookmarks";

export const BookmarkList = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: bookmarks, isLoading, error } = useBookmarks();

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
      <div className="flex flex-col items-center gap-3 p-8 font-sans text-body text-muted">
        <IconBookmark aria-hidden className="size-10 text-subtle" />
        {t("bookmark.empty")}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-0.5 overflow-y-auto p-1.5">
      {bookmarks.map((bookmark) => (
        <MessageLinkCard
          key={`${bookmark.userId}-${bookmark.message.id}`}
          message={bookmark.message}
          workspaceId={workspaceId}
          markedAt={bookmark.createdAt}
        />
      ))}
    </div>
  );
};
