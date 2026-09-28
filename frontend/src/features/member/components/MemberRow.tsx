import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";

import { Avatar } from "#/components/ui/Avatar";
import { focusRing } from "#/components/ui/styles";
import { openPanel } from "#/features/layout/utils/overlaySearch";

type MemberRowProps = {
  userId: string;
  name: string;
  avatarUrl: string | undefined;
  // 名前の下に小さく表示する（メールアドレスなど）
  detail: string;
};

// 押すと右パネルにプロフィールを開くメンバーの行
export const MemberRow = ({ userId, name, avatarUrl, detail }: MemberRowProps) => {
  const navigate = useNavigate();
  return (
    <Button
      onPress={() => {
        void navigate({ search: openPanel({ profile: userId }), to: "." });
      }}
      className={`flex w-full cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-left font-sans text-[13.5px] text-text data-hovered:bg-hover ${focusRing}`}
    >
      <Avatar name={name} src={avatarUrl} size={32} />
      <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
        <span className="truncate">{name}</span>
        <small className="truncate text-xs text-muted">{detail}</small>
      </span>
    </Button>
  );
};
