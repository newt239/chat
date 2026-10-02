import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMembers } from "#/features/member/hooks/useMembers";
import { usePreferences } from "#/hooks/usePreferences";

import { useTypingUsers } from "../hooks/useTypingUsers";

type TypingIndicatorProps = {
  channelId: string;
};

// 入力欄の真上に重ねて出し、表示の有無で入力欄の位置が動かないようにする
export const TypingIndicator = ({ channelId }: TypingIndicatorProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: members } = useMembers(workspaceId);
  const displayName = useDisplayName();
  const userIds = useTypingUsers(channelId);

  if (userIds.length === 0) {
    return null;
  }

  const names = userIds.map((userId) =>
    displayName(
      userId,
      members?.find((member) => member.userId === userId)?.displayName ??
        t("message.typing.someone"),
    ),
  );
  const list = new Intl.ListFormat(locale).format(names.slice(0, 2));

  return (
    <div
      aria-live="polite"
      className="absolute inset-x-0 bottom-full flex h-[18px] items-center gap-1.5 bg-surface px-5 font-sans text-[11.5px] text-muted"
    >
      <span aria-hidden className="inline-flex gap-0.5">
        {[0, 1, 2].map((index) => (
          <span
            key={index}
            className="size-1 animate-pulse rounded-full bg-muted motion-reduce:animate-none"
            style={{ animationDelay: `${index * 200}ms` }}
          />
        ))}
      </span>
      {names.length > 2 ? t("message.typing.many") : t("message.typing.one", { names: list })}
    </div>
  );
};
