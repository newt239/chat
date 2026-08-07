import { createBrowserRouter } from "react-router";

import { registerRouter } from "#/lib/navigation";
import { store } from "#/providers/store";
import { initializeAuthAtom } from "#/providers/store/auth";
import { routeTree } from "#/routes/routeTree";

// 初期ローダー（認証ガード）より先に旧形式トークンを移行する必要がある
store.set(initializeAuthAtom);

export const router = createBrowserRouter(routeTree);

registerRouter(router);
