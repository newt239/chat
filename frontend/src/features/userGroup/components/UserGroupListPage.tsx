import { IconPlus, IconUsers } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState";
import { Link } from "#/components/ui/Link";
import { LinkButton } from "#/components/ui/LinkButton";
import { Skeleton } from "#/components/ui/Skeleton";
import { PageHeader } from "#/features/layout/components/PageHeader";
import { openDialog, openPanel } from "#/features/layout/utils/overlaySearch";

import { useCanManageUserGroups } from "../hooks/useCanManageUserGroups";
import { useUserGroups } from "../hooks/useUserGroups";

// ユーザーグループの一覧。押すと右パネル（モバイルでは全画面）で詳細と編集を開く
export const UserGroupListPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: groups, isLoading } = useUserGroups(workspaceId);
  const canManage = useCanManageUserGroups(workspaceId);

  return (
    <>
      <PageHeader icon={<IconUsers />} title={t("userGroup.pageTitle")}>
        {canManage ? (
          <LinkButton size="sm" to="." search={openDialog({ dialog: "create-group" })}>
            <IconPlus aria-hidden />
            {t("userGroup.create")}
          </LinkButton>
        ) : (
          <span className="text-caption text-muted max-md:hidden">{t("userGroup.adminOnly")}</span>
        )}
      </PageHeader>
      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        {isLoading ? (
          <Skeleton className="h-20 w-full" />
        ) : groups === undefined || groups.length === 0 ? (
          <EmptyState
            icon={<IconUsers />}
            title={t("userGroup.empty")}
            description={canManage ? t("userGroup.emptyHint") : t("userGroup.adminOnly")}
          />
        ) : (
          <ul className="m-0 grid list-none grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-2.5 p-0">
            {groups.map((group) => (
              <li key={group.id}>
                <Link
                  to="."
                  search={openPanel({ group: group.id })}
                  className="flex h-full w-full flex-col items-start gap-1 rounded-xl border border-border bg-surface p-3.5 text-left font-sans no-underline data-hovered:border-border-strong data-hovered:bg-hover"
                >
                  <span className="text-sm font-bold text-accent-text">@{group.name}</span>
                  {group.description !== undefined && group.description.length > 0 && (
                    <span className="line-clamp-2 text-[12.5px] text-muted">
                      {group.description}
                    </span>
                  )}
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </>
  );
};
