import { IconCopy } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { toast } from "#/components/ui/ToastRegion/toast";

type WebhookUrlPanelProps = {
  url: string;
};

// 発行直後にだけ表示する URL と送信例。サーバーはトークンをハッシュでしか持たないため再表示できない
export const WebhookUrlPanel = ({ url }: WebhookUrlPanelProps) => {
  const { t } = useTranslation();
  const copy = () => {
    navigator.clipboard.writeText(url).then(
      () => {
        toast(t("webhook.copied"));
      },
      () => {
        toast(t("webhook.copyFailed"));
      },
    );
  };

  return (
    <div className="flex flex-col gap-2">
      <p className="m-0 rounded-md bg-accent-soft px-3 py-2 text-caption text-accent-text">
        {t("webhook.urlOnce")}
      </p>
      <span className="text-xs font-semibold text-muted">{t("webhook.url")}</span>
      <div className="flex items-center gap-2">
        <code className="min-w-0 flex-1 rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs break-all select-all">
          {url}
        </code>
        <Button size="sm" variant="secondary" onPress={copy}>
          <IconCopy aria-hidden />
          {t("webhook.copy")}
        </Button>
      </div>
      <span className="text-xs font-semibold text-muted">{t("webhook.example")}</span>
      <pre className="m-0 overflow-x-auto rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs whitespace-pre">
        {`curl -X POST -H 'Content-Type: application/json' \\\n  -d '{"text": "Hello"}' \\\n  ${url}`}
      </pre>
    </div>
  );
};
