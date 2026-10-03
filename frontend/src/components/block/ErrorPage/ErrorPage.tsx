import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton/LinkButton";

import type { ErrorComponentProps, NotFoundRouteProps } from "@tanstack/react-router";

// ルートの errorComponent と notFoundComponent の両方に使う
export const ErrorPage = (props: ErrorComponentProps | NotFoundRouteProps) => {
  const { t } = useTranslation();
  const isNotFound = "isNotFound" in props;
  if (!isNotFound) {
    console.error("ルーティングエラー:", props.error);
  }

  return (
    <main className="flex h-full flex-col items-center justify-center gap-3 bg-bg font-sans text-text">
      <h1 className="m-0 text-title">{isNotFound ? "404" : t("shell.error.title")}</h1>
      <p className="m-0 text-muted">
        {t(isNotFound ? "shell.error.notFound" : "shell.error.description")}
      </p>
      <LinkButton to="/app" reloadDocument={!isNotFound}>
        {t("shell.error.backToTop")}
      </LinkButton>
    </main>
  );
};
