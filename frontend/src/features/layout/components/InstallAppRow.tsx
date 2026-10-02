import { IconDownload } from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { navItemClassName } from "#/components/block/NavLink/navTone";
import { cn, focusRing } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { promptInstall, useInstallPrompt } from "#/features/layout/utils/installPrompt";

// 「自分」タブの行。インストールできるときだけ出し、iOS では追加の手順を案内する
export const InstallAppRow = () => {
  const { t } = useTranslation();
  const install = useInstallPrompt();
  if (install === null) {
    return null;
  }

  return (
    <Button
      className={cn(navItemClassName, focusRing)}
      onPress={() => {
        if (install === "ios") {
          toast(t("pwa.install.iosTitle"), { description: t("pwa.install.iosDescription") });
          return;
        }
        void promptInstall(install);
      }}
    >
      <IconDownload aria-hidden />
      {t("pwa.install.action")}
    </Button>
  );
};
