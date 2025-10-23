import { createRouter } from "@tanstack/react-router";

// Import the generated route tree
import { routeTree } from "../routeTree.gen";

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
