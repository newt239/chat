import { createRootRoute } from "@tanstack/react-router";

import { NotFoundPage } from "#/components/block/NotFoundPage/NotFoundPage";
import { RouteErrorBoundary } from "#/components/block/RouteErrorBoundary/RouteErrorBoundary";

export const Route = createRootRoute({
  errorComponent: RouteErrorBoundary,
  notFoundComponent: NotFoundPage,
});
