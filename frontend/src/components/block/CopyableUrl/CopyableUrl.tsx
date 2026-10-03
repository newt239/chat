import { IconCopy } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { copyWithToast } from "#/lib/clipboard";

type CopyableUrlProps = {
  url: string;
};

export const CopyableUrl = ({ url }: CopyableUrlProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex items-center gap-2">
      <code className="min-w-0 flex-1 rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs break-all select-all">
        {url}
      </code>
      <Button
        size="sm"
        variant="secondary"
        onPress={() => {
          void copyWithToast(url, t("ui.copyableUrl.copied"));
        }}
      >
        <IconCopy aria-hidden />
        {t("ui.copyableUrl.copy")}
      </Button>
    </div>
  );
};
