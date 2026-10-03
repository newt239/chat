import { createRootRoute } from "@tanstack/react-router";

import { ErrorPage } from "#/components/block/ErrorPage/ErrorPage";

export const Route = createRootRoute({
  errorComponent: ErrorPage,
  notFoundComponent: ErrorPage,
});
