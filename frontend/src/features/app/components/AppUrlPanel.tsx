import { useTranslation } from "react-i18next";

import { CopyableUrl } from "#/components/block/CopyableUrl/CopyableUrl";

type AppUrlPanelProps = {
  url: string;
};

// 発行直後にだけ表示する着信 Webhook の URL と送信例。サーバーはトークンをハッシュでしか持たないため再表示できない
export const AppUrlPanel = ({ url }: AppUrlPanelProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex flex-col gap-2">
      <p className="m-0 rounded-md bg-accent-soft px-3 py-2 text-caption text-accent-text">
        {t("app.urlOnce")}
      </p>
      <span className="text-xs font-semibold text-muted">{t("app.url")}</span>
      <CopyableUrl url={url} />
      <span className="text-xs font-semibold text-muted">{t("app.example")}</span>
      <pre className="m-0 overflow-x-auto rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs whitespace-pre">
        {`curl -X POST -H 'Content-Type: application/json' \\\n  -d '{"text": "Hello", "channel_id": "<任意>", "thread_id": "<任意>"}' \\\n  ${url}`}
      </pre>
    </div>
  );
};
