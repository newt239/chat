import type { DataRouter } from "react-router";

type NavigableRouter = Pick<DataRouter, "navigate">;

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
