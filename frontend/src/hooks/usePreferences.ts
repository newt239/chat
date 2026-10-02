import { useMemo } from "react";

import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useAtomValue, useSetAtom, useStore } from "jotai";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import { UserService } from "#/gen/chat/v1/user_service_pb";
import { useMe } from "#/hooks/useMe";
import { preferencesFromProto, preferencesToProto } from "#/lib/preferences";
import { storedPreferencesAtom } from "#/providers/store/preferences";

import type { GetMeResponse } from "#/gen/chat/v1/user_service_pb";
import type { Preferences } from "#/providers/store/preferences";

const getMeKey = createConnectQueryKey({ cardinality: "finite", schema: UserService.method.getMe });

/** アカウントの表示設定。ログイン前や読み込み中は端末に残した直近の設定を使う */
export const usePreferences = () => {
  const stored = useAtomValue(storedPreferencesAtom);
  const saved = useMe().data?.preferences;
  // 変換した値を毎回作り直さないよう、保存された設定が変わったときだけ変換する
  return useMemo(() => (saved ? preferencesFromProto(saved) : stored), [saved, stored]);
};

/** 表示設定を変える。preview は画面に反映するだけで、update はアカウントにも保存し、失敗したら元に戻す */
export const useUpdatePreferences = () => {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const store = useStore();
  const setStored = useSetAtom(storedPreferencesAtom);
  const { mutate } = useMutation(UserService.method.updatePreferences);

  const apply = (next: Preferences) => {
    setStored(next);
    queryClient.setQueriesData<GetMeResponse>(
      { queryKey: getMeKey },
      (old) =>
        old?.user && { ...old, user: { ...old.user, preferences: preferencesToProto(next) } },
    );
  };

  // 同じコミットで続けて呼ばれても前の変更を消さないよう、呼ばれた時点の値を読む
  const current = () => {
    const saved = queryClient.getQueriesData<GetMeResponse>({ queryKey: getMeKey })[0]?.[1]?.user
      ?.preferences;
    return saved ? preferencesFromProto(saved) : store.get(storedPreferencesAtom);
  };

  const preview = (patch: Partial<Preferences>) => {
    apply({ ...current(), ...patch });
  };

  const update = (patch: Partial<Preferences>) => {
    const previous = current();
    const next = { ...previous, ...patch };
    apply(next);
    mutate(
      { preferences: preferencesToProto(next) },
      {
        onError: () => {
          apply(previous);
          toast(t("preferences.saveFailed"), { tone: "danger" });
        },
      },
    );
  };

  return { preview, update };
};
