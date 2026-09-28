import { useAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { SegmentedControl } from "#/components/ui/SegmentedControl";
import { Switch } from "#/components/ui/Switch";
import { toast } from "#/components/ui/toast";
import {
  notificationLevels,
  notificationPreferencesAtom,
} from "#/providers/store/notificationPreferences";

import { SettingRow } from "./SettingRow";

export const NotificationSettings = () => {
  const { t } = useTranslation();
  const [preferences, setPreferences] = useAtom(notificationPreferencesAtom);

  return (
    <div className="flex flex-col">
      <SettingRow
        title={t("settings.notifications.desktop")}
        description={t("settings.notifications.desktopDescription")}
      >
        <Switch
          aria-label={t("settings.notifications.desktop")}
          isSelected={preferences.desktop}
          onChange={(desktop) => {
            if (!desktop) {
              setPreferences({ ...preferences, desktop });
              return;
            }
            // 許可を求められるのはユーザー操作の中だけなので、ここで尋ねる
            void Notification.requestPermission().then((permission) => {
              if (permission === "granted") {
                setPreferences({ ...preferences, desktop });
              } else {
                toast(t("settings.notifications.denied"), { tone: "danger" });
              }
            });
          }}
        >
          {null}
        </Switch>
      </SettingRow>
      <SettingRow title={t("settings.notifications.level")} description={null}>
        <SegmentedControl
          label={t("settings.notifications.level")}
          options={notificationLevels.map((value) => ({
            label: t(`settings.notifications.levels.${value}`),
            value,
          }))}
          value={preferences.level}
          onChange={(level) => {
            setPreferences({ ...preferences, level });
          }}
        />
      </SettingRow>
      <p className="m-0 pt-2 text-caption text-muted">{t("settings.notifications.muteHint")}</p>
    </div>
  );
};
