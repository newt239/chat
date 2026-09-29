import {
  IconAt,
  IconBookmark,
  IconChartBar,
  IconFilePencil,
  IconMessages,
  IconShieldCheck,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { ChannelList } from "#/features/channel/components/ChannelList";
import { DMList } from "#/features/dm/components/DMList";
import { UserGroupNavList } from "#/features/userGroup/components/UserGroupNavList";
import { useIsWorkspaceAdmin } from "#/features/workspace/hooks/useIsWorkspaceAdmin";

import { openDialog } from "../utils/overlaySearch";
import { NavLink } from "./NavLink";
import { SidebarSection } from "./SidebarSection";
import { StarredSection } from "./StarredSection";

type NavigationListProps = {
  workspaceId: string;
};

// サイドバーとモバイルのホームで共有する移動先の一覧。配色は親の --nav-* に従う
export const NavigationList = ({ workspaceId }: NavigationListProps) => {
  const { t } = useTranslation();
  const isAdmin = useIsWorkspaceAdmin(workspaceId);
  const navigate = useNavigate();
  const params = { workspaceId };

  return (
    <>
      <nav className="flex min-h-0 flex-1 flex-col overflow-y-auto pb-2">
        <div className="flex flex-col gap-px px-1.5">
          <NavLink to="/app/$workspaceId/threads" params={params}>
            <IconMessages aria-hidden />
            {t("shell.nav.threads")}
          </NavLink>
          <NavLink to="/app/$workspaceId/mentions" params={params}>
            <IconAt aria-hidden />
            {t("shell.nav.mentions")}
          </NavLink>
          <NavLink to="/app/$workspaceId/bookmarks" params={params}>
            <IconBookmark aria-hidden />
            {t("shell.nav.bookmarks")}
          </NavLink>
          <NavLink to="/app/$workspaceId/drafts" params={params}>
            <IconFilePencil aria-hidden />
            {t("draft.page.title")}
          </NavLink>
        </div>
        <StarredSection workspaceId={workspaceId} />
        <SidebarSection
          id="channels"
          title={t("shell.sidebar.channels")}
          onAdd={{
            label: t("shell.sidebar.createChannel"),
            onPress: () => {
              void navigate({ search: openDialog({ dialog: "create-channel" }), to: "." });
            },
          }}
        >
          <ChannelList workspaceId={workspaceId} />
        </SidebarSection>
        <SidebarSection
          id="dms"
          title={t("shell.sidebar.dms")}
          onAdd={{
            label: t("shell.sidebar.createDM"),
            onPress: () => {
              void navigate({ search: openDialog({ dialog: "create-dm" }), to: "." });
            },
          }}
        >
          <DMList workspaceId={workspaceId} />
        </SidebarSection>
        <SidebarSection id="groups" title={t("userGroup.pageTitle")} onAdd={null}>
          <UserGroupNavList workspaceId={workspaceId} />
        </SidebarSection>
        <SidebarSection id="workspace" title={t("shell.sidebar.workspace")} onAdd={null}>
          <NavLink to="/app/$workspaceId/insights" params={params}>
            <IconChartBar aria-hidden />
            {t("shell.nav.insights")}
          </NavLink>
          {isAdmin && (
            <NavLink to="/app/$workspaceId/admin" params={params}>
              <IconShieldCheck aria-hidden />
              {t("shell.nav.admin")}
            </NavLink>
          )}
        </SidebarSection>
      </nav>
    </>
  );
};
