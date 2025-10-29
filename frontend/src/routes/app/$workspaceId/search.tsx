import { createFileRoute } from "@tanstack/react-router";
import { z } from "zod";

import { SearchPage } from "@/features/search/components/SearchPage";

const searchParamsSchema = z.object({
  q: z.string().optional(),
  filter: z.enum(["all", "messages", "channels", "users"]).optional().default("all"),
  page: z.coerce.number().int().min(1).optional().default(1),
});

export type SearchParams = z.infer<typeof searchParamsSchema>;

export const Route = createFileRoute("/app/$workspaceId/search")({
  validateSearch: searchParamsSchema,
  component: SearchPage,
});
