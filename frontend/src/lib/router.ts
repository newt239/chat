import { createRouter } from "@tanstack/react-router";

import { registerRouter } from "#/lib/navigation";
import { routeTree } from "#/routeTree.gen";

export const router = createRouter({ defaultPreload: "intent", routeTree });

registerRouter(router);
