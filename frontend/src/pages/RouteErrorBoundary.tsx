import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton";
import { logger } from "#/lib/logger";

import type { ErrorComponentProps } from "@tanstack/react-router";

export const RouteErrorBoundary = ({ error }: ErrorComponentProps) => {
  const { t } = useTranslation();
  logger.error("ルーティングエラー:", error);

  return (
    <main className="flex h-full flex-col items-center justify-center gap-3 bg-bg font-sans text-text">
      <h1 className="m-0 text-title">{t("shell.error.title")}</h1>
      <p className="m-0 text-muted">{t("shell.error.description")}</p>
      <LinkButton to="/app" reloadDocument>
        {t("shell.error.backToTop")}
      </LinkButton>
    </main>
  );
};
