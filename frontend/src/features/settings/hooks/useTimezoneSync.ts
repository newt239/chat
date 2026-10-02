import { useEffect, useRef } from "react";

import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { useMe } from "#/hooks/useMe";
import { useUpdatePreferences } from "#/hooks/usePreferences";

/** 端末のタイムゾーンがアカウントの設定と違えば更新する。自動更新が無効なら更新するか尋ねる */
export const useTimezoneSync = () => {
  const { t } = useTranslation();
  const { update: updatePreferences } = useUpdatePreferences();
  const data = useMe().data?.preferences;
  const isCheckedRef = useRef(false);

  useEffect(() => {
    if (data === undefined || isCheckedRef.current) {
      return;
    }
    isCheckedRef.current = true;
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (data.timezone === timezone) {
      return;
    }
    // 未設定なら置き換えるものがないので尋ねない
    if (data.timezone === "" || data.timezoneAutoUpdate) {
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
      description: t("preferences.timezone.changedDescription", { current: data.timezone }),
    });
  }, [data, t, updatePreferences]);
};
