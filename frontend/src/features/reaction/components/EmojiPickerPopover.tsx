import { useState } from "react";
import type { ReactElement } from "react";

import { skipToken, useQuery } from "@connectrpc/connect-query";
import { IconMoodPlus } from "@tabler/icons-react";
import { useParams } from "@tanstack/react-router";
import { DialogTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { ResponsivePopover } from "#/components/ui/ResponsivePopover/ResponsivePopover";
import { CustomEmojiForm } from "#/features/customEmoji/components/CustomEmojiForm";
import { toCustomEmojiValue } from "#/features/customEmoji/utils/customEmoji";
import { Permission, PermissionService } from "#/gen/chat/v1/permission_service_pb";

import { EmojiPicker } from "./EmojiPicker";

type EmojiPickerPopoverProps = {
  // React Aria の Button（IconButton など）
  trigger: ReactElement;
  onSelect: (emoji: string) => void;
  // 開いている間はメッセージのツールバーを残すなど、呼び出し側で開閉を知りたいとき
  onOpenChange: ((isOpen: boolean) => void) | null;
  label: string;
  placement: "bottom end" | "top start";
};

export const EmojiPickerPopover = ({
  trigger,
  onSelect,
  onOpenChange,
  label,
  placement,
}: EmojiPickerPopoverProps) => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ strict: false });
  const { data: canCreateEmoji = false } = useQuery(
    PermissionService.method.getPermissions,
    workspaceId === undefined ? skipToken : { workspaceId },
    { select: (res) => res.myPermissions.includes(Permission.CREATE_CUSTOM_EMOJI) },
  );
  const [isOpen, setIsOpen] = useState(false);
  const [isAddOpen, setIsAddOpen] = useState(false);
  const changeOpen = (next: boolean) => {
    setIsOpen(next);
    onOpenChange?.(next);
  };

  return (
    <>
      <DialogTrigger isOpen={isOpen} onOpenChange={changeOpen}>
        {trigger}
        <ResponsivePopover
          aria-label={label}
          placement={placement}
          isOpen={isOpen}
          onOpenChange={changeOpen}
          className="flex flex-col overflow-hidden"
        >
          <EmojiPicker
            onEmojiSelect={(emoji) => {
              onSelect(emoji);
              changeOpen(false);
            }}
          />
          {canCreateEmoji && (
            <div className="border-t border-border p-1.5">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start"
                onPress={() => {
                  changeOpen(false);
                  setIsAddOpen(true);
                }}
              >
                <IconMoodPlus aria-hidden className="size-4" />
                {t("workspace.emoji.add")}
              </Button>
            </div>
          )}
        </ResponsivePopover>
      </DialogTrigger>
      {workspaceId !== undefined && (
        <Dialog isOpen={isAddOpen} onOpenChange={setIsAddOpen} title={t("workspace.emoji.add")}>
          <CustomEmojiForm
            workspaceId={workspaceId}
            onAdded={(name) => {
              setIsAddOpen(false);
              // 登録した絵文字をそのまま使う
              onSelect(toCustomEmojiValue(name));
            }}
          />
        </Dialog>
      )}
    </>
  );
};
