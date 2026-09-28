import { createFileRoute } from "@tanstack/react-router";

import { BookmarksPage } from "#/pages/BookmarksPage";

export const Route = createFileRoute("/app/$workspaceId/bookmarks")({ component: BookmarksPage });
