import type { DataRouter } from "react-router";

/** React のツリー外から遷移するために必要な機能だけを切り出したもの。 */
type NavigableRouter = Pick<DataRouter, "navigate">;

/**
 * React のツリー外(fetch インターセプタや WebSocket クライアント)から遷移するための router 参照。 ルート定義を import せずに済むため、lib →
 * routes → features → lib の循環を避けられる。
 */
let dataRouter: NavigableRouter | null = null;

export const registerRouter = (router: NavigableRouter) => {
  dataRouter = router;
};

export const navigateTo = (path: string) => {
  if (dataRouter === null) {
    return;
  }
  void dataRouter.navigate(path);
};
