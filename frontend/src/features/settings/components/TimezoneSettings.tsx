import { useTranslation } from "react-i18next";

import { ComboBox } from "#/components/ui/ComboBox/ComboBox";
import { Switch } from "#/components/ui/Switch/Switch";
import { usePreferences, useUpdatePreferences } from "#/hooks/usePreferences";

import { SettingRow } from "./SettingRow";

export const TimezoneSettings = () => {
  const { t } = useTranslation();
  const { timezone, timezoneAutoUpdate } = usePreferences();
  const { update: updatePreferences } = useUpdatePreferences();
  // 端末の値が一覧にない別名のこともあるので、保存値は必ず候補に含める
  const options = [...new Set([...Intl.supportedValuesOf("timeZone"), timezone])]
    .filter(Boolean)
    .map((value) => ({ label: value, value }));

  return (
    <>
      <SettingRow
        title={t("preferences.timezone.title")}
        description={t("preferences.timezone.description")}
      >
        <ComboBox
          label={t("preferences.timezone.title")}
          options={options}
          value={timezone || null}
          placeholder={t("preferences.timezone.placeholder")}
          onChange={(value) => {
            updatePreferences({ timezone: value ?? "" });
          }}
          className="w-60 [&>label]:sr-only"
        />
      </SettingRow>
      <SettingRow
        title={t("preferences.timezone.autoUpdate")}
        description={t("preferences.timezone.autoUpdateDescription")}
      >
        <Switch
          aria-label={t("preferences.timezone.autoUpdate")}
          isSelected={timezoneAutoUpdate}
          onChange={(isSelected) => {
            updatePreferences({ timezoneAutoUpdate: isSelected });
          }}
        >
          {null}
        </Switch>
      </SettingRow>
    </>
  );
};
