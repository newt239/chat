import { useState } from "react";

import { IconArrowDown, IconArrowUp, IconDots, IconPencil, IconTrash } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { navIconButtonClassName } from "#/components/ui/styles/navTone";
import { toast } from "#/components/ui/ToastRegion/toast";
import { openDialog } from "#/lib/overlaySearch";

import { useChannelCategoryActions } from "../hooks/useChannelCategories";

import type { ChannelCategory } from "#/gen/chat/v1/channel_category_service_pb";

type ChannelCategoryMenuProps = {
  workspaceId: string;
  category: ChannelCategory;
  // 並び替えに使う全カテゴリの ID
  categoryIds: string[];
};

// 自分で作ったカテゴリの見出しのメニュー。名前の変更・並び替え・削除
export const ChannelCategoryMenu = ({
  workspaceId,
  category,
  categoryIds,
}: ChannelCategoryMenuProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [isDeleting, setIsDeleting] = useState(false);
  const { move, remove } = useChannelCategoryActions(workspaceId);
  const index = categoryIds.indexOf(category.id);

  return (
    <>
      <Menu
        placement="bottom start"
        trigger={
          <IconButton
            label={t("channel.category.menu", { name: category.name })}
            className={navIconButtonClassName}
          >
            <IconDots />
          </IconButton>
        }
      >
        <MenuItem
          icon={<IconPencil />}
          onAction={() => {
            void navigate({
              search: openDialog({ category: category.id, dialog: "edit-category" }),
              to: ".",
            });
          }}
        >
          {t("channel.category.rename")}
        </MenuItem>
        <MenuItem
          icon={<IconArrowUp />}
          isDisabled={index <= 0}
          onAction={() => {
            move(categoryIds, index, index - 1);
          }}
        >
          {t("channel.category.moveUp")}
        </MenuItem>
        <MenuItem
          icon={<IconArrowDown />}
          isDisabled={index === categoryIds.length - 1}
          onAction={() => {
            move(categoryIds, index, index + 1);
          }}
        >
          {t("channel.category.moveDown")}
        </MenuItem>
        <MenuSeparator />
        <MenuItem
          icon={<IconTrash />}
          tone="danger"
          onAction={() => {
            setIsDeleting(true);
          }}
        >
          {t("channel.category.delete")}
        </MenuItem>
      </Menu>
      <AlertDialog
        isOpen={isDeleting}
        onOpenChange={setIsDeleting}
        title={t("channel.category.deleteConfirm", { name: category.name })}
        confirmLabel={t("channel.category.delete")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          remove.mutate(
            { categoryId: category.id },
            {
              onSuccess: () => {
                setIsDeleting(false);
                toast(t("channel.category.deleted"));
              },
            },
          );
        }}
      >
        {t("channel.category.deleteDescription")}
      </AlertDialog>
    </>
  );
};
