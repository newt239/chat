import { useTranslation } from "react-i18next";

type GroupAvatarProps = {
  // 自分を含めた人数
  count: number;
  size?: number;
};

// グループ DM のアイコン。顔を並べる代わりに人数を表示する
export const GroupAvatar = ({ count, size = 32 }: GroupAvatarProps) => {
  const { t } = useTranslation();
  return (
    <span
      role="img"
      aria-label={t("ui.avatar.groupMembers", { count })}
      className="inline-grid shrink-0 place-items-center rounded-[28%] border border-border bg-sunken font-mono leading-none font-bold text-muted select-none"
      style={{ fontSize: Math.max(9, Math.round(size * 0.42)), height: size, width: size }}
    >
      <span aria-hidden>{count}</span>
    </span>
  );
};
