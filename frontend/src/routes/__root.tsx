import { createRootRoute } from "@tanstack/react-router";

import { NotFoundPage } from "#/pages/NotFoundPage";
import { RouteErrorBoundary } from "#/pages/RouteErrorBoundary";

export const Route = createRootRoute({
  errorComponent: RouteErrorBoundary,
  notFoundComponent: NotFoundPage,
});
