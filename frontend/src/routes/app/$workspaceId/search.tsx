import { createFileRoute } from "@tanstack/react-router";

import { SearchPage } from "#/features/search/components/SearchPage";
import { searchQuerySchema } from "#/features/search/schemas";

export const Route = createFileRoute("/app/$workspaceId/search")({
  component: SearchPage,
  validateSearch: searchQuerySchema,
});
