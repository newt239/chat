import type { NavigateOptions } from "@tanstack/react-router";

// router はルートを通じて session などを読み込み、静的に import すると循環するため遅延させる
export const navigateTo = (options: NavigateOptions) => {
  void import("#/lib/router").then(({ router }) => router.navigate(options));
};
