import { IconCopy } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { toast } from "#/components/ui/ToastRegion/toast";

type CopyableUrlProps = {
  url: string;
};

export const CopyableUrl = ({ url }: CopyableUrlProps) => {
  const { t } = useTranslation();

  const copy = () => {
    navigator.clipboard.writeText(url).then(
      () => {
        toast(t("ui.copyableUrl.copied"));
      },
      () => {
        toast(t("ui.copyableUrl.copyFailed"));
      },
    );
  };

  return (
    <div className="flex items-center gap-2">
      <code className="min-w-0 flex-1 rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs break-all select-all">
        {url}
      </code>
      <Button size="sm" variant="secondary" onPress={copy}>
        <IconCopy aria-hidden />
        {t("ui.copyableUrl.copy")}
      </Button>
    </div>
  );
};
