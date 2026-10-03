import { IconCheck, IconDots, IconFolderPlus, IconHash } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { MenuSection } from "#/components/ui/MenuSection/MenuSection";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { navIconButtonClassName } from "#/components/ui/styles/navTone";
import { usePreferences, useUpdatePreferences } from "#/hooks/usePreferences";
import { openDialog } from "#/lib/overlaySearch";
import { channelSortOrders } from "#/providers/store/preferences";

type ChannelSectionMenuProps = {
  workspaceId: string;
};

// サイドバーの「チャンネル」見出しのメニュー。並び順の切り替え・カテゴリの作成・チャンネルの一覧
export const ChannelSectionMenu = ({ workspaceId }: ChannelSectionMenuProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { channelSortOrder } = usePreferences();
  const { update: updatePreferences } = useUpdatePreferences();

  return (
    <Menu
      placement="bottom start"
      trigger={
        <IconButton label={t("channel.sectionMenu")} className={navIconButtonClassName}>
          <IconDots />
        </IconButton>
      }
    >
      <MenuSection title={t("channel.sort.title")}>
        {channelSortOrders.map((order) => (
          <MenuItem
            key={order}
            icon={<IconCheck className={order === channelSortOrder ? "" : "invisible"} />}
            onAction={() => {
              updatePreferences({ channelSortOrder: order });
            }}
          >
            {t(`channel.sort.${order}`)}
          </MenuItem>
        ))}
      </MenuSection>
      <MenuSeparator />
      <MenuItem
        icon={<IconFolderPlus />}
        onAction={() => {
          void navigate({ search: openDialog({ dialog: "create-category" }), to: "." });
        }}
      >
        {t("channel.category.create")}
      </MenuItem>
      <MenuItemLink
        to="/app/$workspaceId/browse-channels"
        params={{ workspaceId }}
        icon={<IconHash />}
      >
        {t("channel.browse.title")}
      </MenuItemLink>
    </Menu>
  );
};
