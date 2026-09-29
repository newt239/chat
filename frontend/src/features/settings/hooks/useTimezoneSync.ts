import { useEffect, useRef } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import { UserService } from "#/gen/chat/v1/user_service_pb";

import { useUpdatePreferences } from "./usePreferences";

/** 端末のタイムゾーンがアカウントの設定と違えば更新する。自動更新が無効なら更新するか尋ねる。 useSyncPreferences より後に呼ぶ */
export const useTimezoneSync = () => {
  const { t } = useTranslation();
  const updatePreferences = useUpdatePreferences();
  const { data } = useQuery(
    UserService.method.getMe,
    {},
    { select: (res) => res.user?.preferences },
  );
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
