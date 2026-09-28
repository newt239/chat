import { IconBell, IconKeyboard, IconKey, IconLanguage, IconPalette } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Dialog } from "#/components/ui/Dialog";
import { Link } from "#/components/ui/Link";
import { settingsSections } from "#/features/layout/schemas";
import { closeDialog, openDialog } from "#/features/layout/utils/overlaySearch";
import { workspaceRoute } from "#/features/layout/utils/workspaceRoute";
import { useIsMobile } from "#/lib/useMediaQuery";

import { AccountSettings } from "./AccountSettings";
import { DisplaySettings } from "./DisplaySettings";
import { NotificationSettings } from "./NotificationSettings";
import { ShortcutSettings } from "./ShortcutSettings";
import { ThemeSettings } from "./ThemeSettings";

import type { SettingsSection } from "#/features/layout/schemas";

const sectionIcons: Record<SettingsSection, typeof IconKey> = {
  account: IconKey,
  display: IconLanguage,
  notifications: IconBell,
  shortcuts: IconKeyboard,
  theme: IconPalette,
};

const sectionBodies: Record<SettingsSection, () => React.JSX.Element> = {
  account: AccountSettings,
  display: DisplaySettings,
  notifications: NotificationSettings,
  shortcuts: ShortcutSettings,
  theme: ThemeSettings,
};

// ?settings= で開く。デスクトップでは左に項目を並べたモーダル、モバイルでは項目ごとの全画面ページ
export const SettingsDialog = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const section = workspaceRoute.useSearch({ select: (search) => search.settings });
  const current = section ?? "account";
  const Body = sectionBodies[current];

  return (
    <Dialog
      isOpen={section !== undefined}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          void navigate({ search: closeDialog, to: "." });
        }
      }}
      title={isMobile ? t(`settings.sections.${current}`) : t("settings.title")}
      size="lg"
    >
      <div className="flex min-h-[min(520px,70vh)] gap-6">
        {!isMobile && (
          <nav aria-label={t("settings.title")} className="flex w-44 shrink-0 flex-col gap-px">
            {settingsSections.map((name) => {
              const Icon = sectionIcons[name];
              return (
                <Link
                  key={name}
                  aria-current={name === current ? "page" : undefined}
                  to="."
                  search={openDialog({ settings: name })}
                  replace
                  className="flex h-8 items-center gap-2 rounded-md px-2.5 text-[13.5px] text-muted no-underline data-hovered:bg-hover data-hovered:text-text aria-[current=page]:bg-accent-soft aria-[current=page]:font-semibold aria-[current=page]:text-accent-text [&_svg]:size-4"
                >
                  <Icon aria-hidden />
                  {t(`settings.sections.${name}`)}
                </Link>
              );
            })}
          </nav>
        )}
        <section className="flex min-w-0 flex-1 flex-col gap-3">
          {!isMobile && <h3 className="m-0 text-title">{t(`settings.sections.${current}`)}</h3>}
          <Body />
        </section>
      </div>
    </Dialog>
  );
};
