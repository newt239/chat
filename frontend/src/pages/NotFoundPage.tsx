import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton";

export const NotFoundPage = () => {
  const { t } = useTranslation();
  return (
    <main className="flex h-full flex-col items-center justify-center gap-3 bg-bg font-sans text-text">
      <h1 className="m-0 text-title">404</h1>
      <p className="m-0 text-muted">{t("shell.error.notFound")}</p>
      <LinkButton to="/app">{t("shell.error.backToTop")}</LinkButton>
    </main>
  );
};
