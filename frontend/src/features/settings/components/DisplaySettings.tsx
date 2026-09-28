import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { SegmentedControl } from "#/components/ui/SegmentedControl";
import { preferencesAtom } from "#/providers/store/preferences";

import { useUpdatePreferences } from "../hooks/usePreferences";
import { SettingRow } from "./SettingRow";

export const DisplaySettings = () => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const updatePreferences = useUpdatePreferences();

  return (
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
  );
};
