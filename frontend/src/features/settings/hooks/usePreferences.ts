import { useEffect } from "react";

import { createConnectQueryKey, useMutation, useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useSetAtom, useStore } from "jotai";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { preferencesAtom } from "#/providers/store/preferences";

import { preferencesFromProto, preferencesToProto } from "../utils/preferences";

import type { GetMeResponse } from "#/gen/chat/v1/user_service_pb";
import type { Preferences } from "#/providers/store/preferences";

/** アカウントに保存された表示設定を読み込み、端末の設定を上書きする */
export const useSyncPreferences = () => {
  const setPreferences = useSetAtom(preferencesAtom);
  const { data } = useQuery(
    UserService.method.getMe,
    {},
    { select: (res) => res.user?.preferences },
  );

  useEffect(() => {
    if (data) {
      setPreferences(preferencesFromProto(data));
    }
  }, [data, setPreferences]);
};

/** 表示設定をすぐに反映し、アカウントに保存する。保存に失敗したら元に戻す */
export const useUpdatePreferences = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const store = useStore();
  const setPreferences = useSetAtom(preferencesAtom);
  const { mutate } = useMutation(UserService.method.updatePreferences);

  return (patch: Partial<Preferences>) => {
    // 同じコミットで読み込んだアカウントの設定も含めるため、呼ばれた時点の値を読む
    const previous = store.get(preferencesAtom);
    const next = { ...previous, ...patch };
    setPreferences(next);
    mutate(
      { preferences: preferencesToProto(next) },
      {
        onError: () => {
          setPreferences(previous);
          toast(t("preferences.saveFailed"), { tone: "danger" });
        },
        // 古いキャッシュで useSyncPreferences が設定を巻き戻さないようにする
        onSuccess: (res) => {
          queryClient.setQueriesData<GetMeResponse>(
            {
              queryKey: createConnectQueryKey({
                cardinality: "finite",
                schema: UserService.method.getMe,
              }),
            },
            (old) => old?.user && { ...old, user: { ...old.user, preferences: res.preferences } },
          );
        },
      },
    );
  };
};
