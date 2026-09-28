import type { router } from "#/lib/router";

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
