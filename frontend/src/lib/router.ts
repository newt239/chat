import { createBrowserRouter } from "react-router";

import { registerRouter } from "#/lib/navigation";
import { store } from "#/providers/store";
import { initializeAuthAtom } from "#/providers/store/auth";
import { routeTree } from "#/routes/routeTree";

// CreateBrowserRouter は生成直後に初期ローダー（＝認証ガード）を走らせる。
// 旧形式トークンの移行はそれより前に完了している必要があるため、モジュールスコープで実行する。
store.set(initializeAuthAtom);

export const router = createBrowserRouter(routeTree);

registerRouter(router);
