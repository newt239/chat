import { useAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { SegmentedControl } from "#/components/ui/SegmentedControl/SegmentedControl";
import { Switch } from "#/components/ui/Switch/Switch";
import { toast } from "#/components/ui/ToastRegion/toast";
import { notificationPreferencesAtom } from "#/features/notification/atoms";
import { usePushNotifications } from "#/features/notification/hooks/usePushNotifications";
import {
  isNotificationSupported,
  requestNotificationPermission,
} from "#/features/notification/utils/notify";
import { usePreferences, useUpdatePreferences } from "#/hooks/usePreferences";
import { notificationLevels } from "#/providers/store/preferences";

import { SettingRow } from "./SettingRow";

export const NotificationSettings = () => {
  const { t } = useTranslation();
  const [device, setDevice] = useAtom(notificationPreferencesAtom);
  const { notificationLevel } = usePreferences();
  const { update: updatePreferences } = useUpdatePreferences();
  const push = usePushNotifications();

  return (
    <div className="flex flex-col">
      <SettingRow title={t("settings.notifications.level")} description={null}>
        <SegmentedControl
          label={t("settings.notifications.level")}
          options={notificationLevels.map((value) => ({
            label: t(`settings.notifications.levels.${value}`),
            value,
          }))}
          value={notificationLevel}
          onChange={(level) => {
            updatePreferences({ notificationLevel: level });
          }}
        />
      </SettingRow>
      {isNotificationSupported() && (
        <SettingRow
          title={t("settings.notifications.desktop")}
          description={t("settings.notifications.desktopDescription")}
        >
          <Switch
            aria-label={t("settings.notifications.desktop")}
            isSelected={device.desktop}
            onChange={(desktop) => {
              if (!desktop) {
                setDevice((prev) => ({ ...prev, desktop }));
                return;
              }
              // 許可を求められるのはユーザー操作の中だけなので、ここで尋ねる
              void requestNotificationPermission().then((granted) => {
                if (granted) {
                  setDevice((prev) => ({ ...prev, desktop }));
                } else {
                  toast(t("settings.notifications.denied"), { tone: "danger" });
                }
              });
            }}
          >
            {null}
          </Switch>
        </SettingRow>
      )}
      {push.supported && (
        <SettingRow
          title={t("settings.notifications.push")}
          description={t("settings.notifications.pushDescription")}
        >
          <Switch
            aria-label={t("settings.notifications.push")}
            isSelected={push.enabled}
            onChange={(enabled) => {
              if (!enabled) {
                void push.disable();
                return;
              }
              void (async () => {
                try {
                  if (!(await push.enable())) {
                    toast(t("settings.notifications.denied"), { tone: "danger" });
                  }
                } catch (error) {
                  console.warn("プッシュ通知を登録できませんでした", error);
                  toast(t("settings.notifications.pushFailed"), { tone: "danger" });
                }
              })();
            }}
          >
            {null}
          </Switch>
        </SettingRow>
      )}
      <p className="m-0 pt-2 text-caption text-muted">{t("settings.notifications.muteHint")}</p>
    </div>
  );
};
