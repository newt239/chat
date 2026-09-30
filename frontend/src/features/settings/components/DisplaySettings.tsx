import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { SegmentedControl } from "#/components/ui/SegmentedControl/SegmentedControl";
import { Switch } from "#/components/ui/Switch/Switch";
import { channelSortOrders, preferencesAtom } from "#/providers/store/preferences";

import { useUpdatePreferences } from "../hooks/usePreferences";
import { SettingRow } from "./SettingRow";
import { TimezoneSettings } from "./TimezoneSettings";

export const DisplaySettings = () => {
  const { t } = useTranslation();
  const { channelSortOrder, hideJoinMessages, locale } = useAtomValue(preferencesAtom);
  const updatePreferences = useUpdatePreferences();

  return (
    <div className="flex flex-col">
      <SettingRow title={t("preferences.locale.title")} description="Language">
        <SegmentedControl
          label={t("preferences.locale.title")}
          options={(["ja", "en"] as const).map((value) => ({
            label: t(`preferences.locale.${value}`),
            value,
          }))}
          value={locale}
          onChange={(value) => {
            updatePreferences({ locale: value });
          }}
        />
      </SettingRow>
      <TimezoneSettings />
      <SettingRow
        title={t("preferences.channelSort.title")}
        description={t("preferences.channelSort.description")}
      >
        <SegmentedControl
          label={t("preferences.channelSort.title")}
          options={channelSortOrders.map((value) => ({
            label: t(`channel.sort.${value}`),
            value,
          }))}
          value={channelSortOrder}
          onChange={(value) => {
            updatePreferences({ channelSortOrder: value });
          }}
        />
      </SettingRow>
      <SettingRow
        title={t("preferences.joinMessages.title")}
        description={t("preferences.joinMessages.description")}
      >
        <Switch
          aria-label={t("preferences.joinMessages.title")}
          isSelected={!hideJoinMessages}
          onChange={(isSelected) => {
            updatePreferences({ hideJoinMessages: !isSelected });
          }}
        >
          {null}
        </Switch>
      </SettingRow>
    </div>
  );
};
