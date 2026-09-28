import { IconBell, IconKeyboard, IconKey, IconLanguage, IconPalette } from "@tabler/icons-react";
import { useAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Dialog } from "#/components/ui/Dialog";
import { focusRing } from "#/components/ui/styles";
import { useIsMobile } from "#/lib/useMediaQuery";
import { settingsSectionAtom, settingsSections } from "#/providers/store/ui";

import { AccountSettings } from "./AccountSettings";
import { DisplaySettings } from "./DisplaySettings";
import { NotificationSettings } from "./NotificationSettings";
import { ShortcutSettings } from "./ShortcutSettings";
import { ThemeSettings } from "./ThemeSettings";

import type { SettingsSection } from "#/providers/store/ui";

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

// デスクトップでは左に項目を並べたモーダル、モバイルでは項目ごとの全画面ページ
export const SettingsDialog = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [section, setSection] = useAtom(settingsSectionAtom);
  const current = section ?? "account";
  const Body = sectionBodies[current];

  return (
    <Dialog
      isOpen={section !== null}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          setSection(null);
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
                <Button
                  key={name}
                  aria-current={name === current ? "page" : undefined}
                  onPress={() => {
                    setSection(name);
                  }}
                  className={`flex h-8 cursor-pointer items-center gap-2 rounded-md px-2.5 text-left text-[13.5px] text-muted data-hovered:bg-hover data-hovered:text-text aria-[current=page]:bg-accent-soft aria-[current=page]:font-semibold aria-[current=page]:text-accent-text [&_svg]:size-4 ${focusRing}`}
                >
                  <Icon aria-hidden />
                  {t(`settings.sections.${name}`)}
                </Button>
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
