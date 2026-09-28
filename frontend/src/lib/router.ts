import { createRouter } from "@tanstack/react-router";

import { registerRouter } from "#/lib/navigation";
import { store } from "#/providers/store";
import { initializeAuthAtom } from "#/providers/store/auth";
import { routeTree } from "#/routeTree.gen";

// 初回の beforeLoad（認証ガード）より先に旧形式トークンを移行する必要がある
store.set(initializeAuthAtom);

export const router = createRouter({ defaultPreload: "intent", routeTree });

registerRouter(router);
