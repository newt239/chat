import {
  IconAt,
  IconBookmark,
  IconFilePencil,
  IconMessages,
  IconPlus,
  IconShieldCheck,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { NavLink } from "#/components/block/NavLink/NavLink";
import { ChannelCategoryMenu } from "#/features/channel/components/ChannelCategoryMenu";
import { ChannelList } from "#/features/channel/components/ChannelList";
import { ChannelSectionMenu } from "#/features/channel/components/ChannelSectionMenu";
import { useChannelCategories } from "#/features/channel/hooks/useChannelCategories";
import { DMList } from "#/features/dm/components/DMList";
import { UserGroupNavList } from "#/features/userGroup/components/UserGroupNavList";
import { useIsWorkspaceAdmin } from "#/hooks/useIsWorkspaceAdmin";
import { openDialog } from "#/lib/overlaySearch";

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
  const { data: categories = [] } = useChannelCategories(workspaceId);
  const categoryIds = categories.map((category) => category.id);
  const params = { workspaceId };

  return (
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
      {categories.map((category) => (
        <SidebarSection
          key={category.id}
          id={`category:${category.id}`}
          title={category.name}
          onAdd={null}
          menu={
            <ChannelCategoryMenu
              workspaceId={workspaceId}
              category={category}
              categoryIds={categoryIds}
            />
          }
        >
          <ChannelList workspaceId={workspaceId} categoryId={category.id} />
        </SidebarSection>
      ))}
      <SidebarSection
        id="channels"
        title={t("shell.sidebar.channels")}
        menu={<ChannelSectionMenu workspaceId={workspaceId} />}
        onAdd={{
          label: t("shell.sidebar.createChannel"),
          onPress: () => {
            void navigate({ search: openDialog({ dialog: "create-channel" }), to: "." });
          },
        }}
      >
        <ChannelList workspaceId={workspaceId} categoryId={null} />
        <NavLink to="/app/$workspaceId/browse-channels" params={params}>
          <IconPlus aria-hidden />
          {t("shell.sidebar.browseChannels")}
        </NavLink>
      </SidebarSection>
      <SidebarSection
        id="dms"
        title={t("shell.sidebar.dms")}
        menu={null}
        onAdd={{
          label: t("shell.sidebar.createDM"),
          onPress: () => {
            void navigate({ search: openDialog({ dialog: "create-dm" }), to: "." });
          },
        }}
      >
        <DMList workspaceId={workspaceId} />
      </SidebarSection>
      <SidebarSection
        id="groups"
        title={t("userGroup.pageTitle")}
        menu={null}
        onAdd={
          isAdmin
            ? {
                label: t("userGroup.createLabel"),
                onPress: () => {
                  void navigate({ search: openDialog({ dialog: "create-group" }), to: "." });
                },
              }
            : null
        }
      >
        <UserGroupNavList workspaceId={workspaceId} />
      </SidebarSection>
      {isAdmin && (
        <SidebarSection
          id="workspace"
          title={t("shell.sidebar.workspace")}
          onAdd={null}
          menu={null}
        >
          <NavLink to="/app/$workspaceId/admin" params={params}>
            <IconShieldCheck aria-hidden />
            {t("shell.nav.admin")}
          </NavLink>
        </SidebarSection>
      )}
    </nav>
  );
};
