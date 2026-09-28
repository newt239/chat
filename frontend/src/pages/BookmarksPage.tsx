import { IconBookmark } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { BookmarkList } from "#/features/bookmark/components/BookmarkList";
import { PageHeader } from "#/features/layout/components/PageHeader";

export const BookmarksPage = () => {
  const { t } = useTranslation();
  return (
    <>
      <PageHeader icon={<IconBookmark />} title={t("shell.nav.bookmarks")} />
      <div className="min-h-0 flex-1 overflow-y-auto">
        <BookmarkList />
      </div>
    </>
  );
};
