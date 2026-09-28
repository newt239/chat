import type { AnyRouter, NavigateOptions } from "@tanstack/react-router";

let registeredRouter: Pick<AnyRouter, "navigate"> | null = null;

export const registerRouter = (router: Pick<AnyRouter, "navigate">) => {
  registeredRouter = router;
};

export const navigateTo = (options: NavigateOptions) => {
  if (registeredRouter === null) {
    return;
  }
  void registeredRouter.navigate(options);
};
