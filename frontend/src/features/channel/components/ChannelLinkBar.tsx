import { IconEdit, IconLink, IconPlus, IconTrash } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { ContextMenu } from "#/components/ui/ContextMenu/ContextMenu";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuItemLink } from "#/components/ui/MenuItemLink/MenuItemLink";
import { focusRing } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { openDialog } from "#/features/layout/utils/overlaySearch";

import { useChannelLinkActions, useChannelLinks } from "../hooks/useChannelLinks";

type ChannelLinkBarProps = {
  channelId: string;
};

// チャンネルヘッダーの下に並べる関連リンク。編集できる人は右クリックで編集・削除し、+ で追加する
export const ChannelLinkBar = ({ channelId }: ChannelLinkBarProps) => {
  const { t } = useTranslation();
  const { data } = useChannelLinks(channelId);
  const { remove } = useChannelLinkActions(channelId);
  const navigate = useNavigate();
  const links = data?.links ?? [];

  if (links.length === 0) {
    return null;
  }

  const linkClassName = `inline-flex h-6 shrink-0 cursor-pointer items-center gap-1.5 rounded-md px-2 text-label font-normal whitespace-nowrap text-text no-underline data-hovered:bg-hover [&_svg]:size-3.5 [&_svg]:text-muted ${focusRing}`;

  return (
    <nav
      aria-label={t("shell.channel.links")}
      className="flex h-8 shrink-0 items-center gap-0.5 overflow-x-auto border-b border-border pr-2.5 pl-3.5 [scrollbar-width:none]"
    >
      {links.map((link) => {
        const anchor = (
          <Link href={link.url} target="_blank" rel="noopener noreferrer" className={linkClassName}>
            <IconLink aria-hidden />
            {link.title}
          </Link>
        );
        return data?.canEdit ? (
          <ContextMenu
            key={link.id}
            aria-label={t("channel.links.menu", { title: link.title })}
            menu={
              <>
                <MenuItemLink
                  icon={<IconEdit />}
                  to="."
                  search={openDialog({ dialog: "edit-link", link: link.id })}
                >
                  {t("channel.links.edit")}
                </MenuItemLink>
                <MenuItem
                  icon={<IconTrash />}
                  tone="danger"
                  onAction={() => {
                    remove.mutate(
                      { linkId: link.id },
                      {
                        onSuccess: () => {
                          toast(t("channel.links.deleted"));
                        },
                      },
                    );
                  }}
                >
                  {t("common.delete")}
                </MenuItem>
              </>
            }
          >
            {anchor}
          </ContextMenu>
        ) : (
          <span key={link.id}>{anchor}</span>
        );
      })}
      {data?.canEdit && (
        <IconButton
          label={t("channel.links.addTitle")}
          className="size-6 [&_svg]:size-3.5"
          onPress={() => {
            void navigate({ search: openDialog({ dialog: "add-link" }), to: "." });
          }}
        >
          <IconPlus />
        </IconButton>
      )}
    </nav>
  );
};
