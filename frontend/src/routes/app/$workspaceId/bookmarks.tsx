import { createFileRoute } from "@tanstack/react-router";

import { BookmarksPage } from "#/features/bookmark/components/BookmarksPage";

export const Route = createFileRoute("/app/$workspaceId/bookmarks")({ component: BookmarksPage });
