import { IconSettings } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { SettingsLayout } from "#/components/block/SettingsLayout/SettingsLayout";
import { settingsNavLinkClassName } from "#/components/block/SettingsLayout/settingsNavLinkClassName";
import { Link } from "#/components/ui/Link/Link";

import { findSettingsSection, settingsSections } from "../schemas";
import { settingsSectionIcons } from "../utils/sectionIcons";
import { AccountSettings } from "./AccountSettings";
import { DisplaySettings } from "./DisplaySettings";
import { NotificationSettings } from "./NotificationSettings";
import { ShortcutSettings } from "./ShortcutSettings";
import { ThemeSettings } from "./ThemeSettings";

import type { SettingsSection } from "../schemas";

const sectionBodies: Record<SettingsSection, () => React.JSX.Element> = {
  account: AccountSettings,
  display: DisplaySettings,
  notifications: NotificationSettings,
  shortcuts: ShortcutSettings,
  theme: ThemeSettings,
};

const settingsRoute = getRouteApi("/app/$workspaceId/settings/{-$section}");

export const SettingsPage = () => {
  const { t } = useTranslation();
  const { section, workspaceId } = settingsRoute.useParams();
  const current = findSettingsSection(section) ?? "account";
  const Body = sectionBodies[current];

  return (
    <SettingsLayout
      icon={<IconSettings />}
      title={t("settings.title")}
      sectionTitle={t(`settings.sections.${current}`)}
      nav={settingsSections.map((name) => {
        const Icon = settingsSectionIcons[name];
        return (
          <Link
            className={settingsNavLinkClassName}
            key={name}
            to="/app/$workspaceId/settings/{-$section}"
            params={{ section: name, workspaceId }}
            replace
          >
            <Icon aria-hidden />
            {t(`settings.sections.${name}`)}
          </Link>
        );
      })}
    >
      <Body />
    </SettingsLayout>
  );
};
