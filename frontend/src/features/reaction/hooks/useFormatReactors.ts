import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { usePreferences } from "#/hooks/usePreferences";
import { myUserIdAtom } from "#/providers/store/auth";

import type { UserSummary } from "#/gen/chat/v1/user_pb";

const MAX_NAMES = 4;

// 自分を「あなた」として先頭に置き、多いときは先頭の 3 人と残りの人数にまとめる
export const useFormatReactors = () => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const currentUserId = useAtomValue(myUserIdAtom);
  const listFormat = new Intl.ListFormat(locale);
  const displayName = useDisplayName();

  return (users: UserSummary[]) => {
    const names = users
      .toSorted((a, b) => Number(b.id === currentUserId) - Number(a.id === currentUserId))
      .map((user) =>
        user.id === currentUserId
          ? t("reaction.names.you")
          : displayName(user.id, user.displayName),
      );
    if (names.length <= MAX_NAMES) {
      return listFormat.format(names);
    }
    return t("reaction.names.others", {
      count: names.length - (MAX_NAMES - 1),
      names: listFormat.format(names.slice(0, MAX_NAMES - 1)),
    });
  };
};
