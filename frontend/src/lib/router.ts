import { createRouter } from "@tanstack/react-router";

// Import the generated route tree
import { routeTree } from "@/routes/routeTree.gen";

export const router = createRouter({ routeTree });
