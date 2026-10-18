import { IconSearch } from "@tabler/icons-react";
import { Keyboard } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Link } from "#/components/ui/Link/Link";
import { surfaceNavTone } from "#/components/ui/styles/navTone";
import { WorkspaceMenu } from "#/features/workspace/components/WorkspaceMenu";
import { openDialog } from "#/lib/overlaySearch";

import { AccountMenu } from "./AccountMenu";
import { NextUnreadButton } from "./NextUnreadButton";

type TopBarProps = {
  workspaceId: string;
};

// デスクトップの最上段。ワークスペースの切り替え・移動・次の未読・自分のメニューを置く
export const TopBar = ({ workspaceId }: TopBarProps) => {
  const { t } = useTranslation();
  return (
    <header className={`flex h-12 shrink-0 items-center gap-3 px-2 ${surfaceNavTone}`}>
      <div className="flex w-60 min-w-0 shrink-0">
        <WorkspaceMenu workspaceId={workspaceId} />
      </div>
      <div className="flex min-w-0 flex-1 justify-center">
        <Link
          to="."
          search={openDialog({ dialog: "switcher" })}
          className="flex h-8.5 w-full max-w-115 items-center gap-2 rounded-full border border-border bg-surface px-3.5 text-body-sm text-muted no-underline data-hovered:border-border-strong [&_svg]:size-4"
        >
          <IconSearch aria-hidden />
          <span className="min-w-0 flex-1 truncate">{t("shell.switcher.open")}</span>
          <Keyboard className="rounded-sm border border-border px-1.5 font-mono text-caption">
            ⌘K
          </Keyboard>
        </Link>
      </div>
      <div className="flex w-60 shrink-0 items-center justify-end gap-2">
        <NextUnreadButton workspaceId={workspaceId} />
        <AccountMenu />
      </div>
    </header>
  );
};
