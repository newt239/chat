import { useEffect, useEffectEvent } from "react";

import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { useMe } from "#/hooks/useMe";
import { useUpdatePreferences } from "#/hooks/usePreferences";

// 読み込んだときに一度だけ、端末のタイムゾーンがアカウントの設定と違えば更新する。自動更新が無効なら更新するか尋ねる
export const useTimezoneSync = () => {
  const { t } = useTranslation();
  const { update: updatePreferences } = useUpdatePreferences();
  const preferences = useMe().data?.preferences;
  const isLoaded = preferences !== undefined;

  const sync = useEffectEvent(() => {
    if (preferences === undefined) {
      return;
    }
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (preferences.timezone === timezone) {
      return;
    }
    // 未設定なら置き換えるものがないので尋ねない
    if (preferences.timezone === "" || preferences.timezoneAutoUpdate) {
      updatePreferences({ timezone });
      return;
    }
    toast(t("preferences.timezone.changed", { timezone }), {
      action: {
        label: t("preferences.timezone.update"),
        onAction: () => {
          updatePreferences({ timezone });
        },
      },
      description: t("preferences.timezone.changedDescription", { current: preferences.timezone }),
    });
  });

  useEffect(() => {
    if (isLoaded) {
      sync();
    }
  }, [isLoaded]);
};
