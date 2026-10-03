import { IconPlus, IconX } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { toast } from "#/components/ui/ToastRegion/toast";
import { openDialog } from "#/lib/overlaySearch";

import { useAppActions, useApps, useChannelApps } from "../hooks/useApps";
import { AppRow } from "./AppRow";

type ChannelAppsSectionProps = {
  workspaceId: string;
  channelId: string;
};

// チャンネル情報のアプリ一覧。管理できるアプリを追加・外したり、新しく作ったりできる
export const ChannelAppsSection = ({ workspaceId, channelId }: ChannelAppsSectionProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data } = useChannelApps(channelId);
  const { data: all } = useApps(workspaceId);
  const { addToChannel, removeFromChannel } = useAppActions();
  const joined = data?.apps ?? [];
  const addable = (all?.apps ?? []).filter(
    (app) => app.canManage && !joined.some((candidate) => candidate.id === app.id),
  );

  return (
    <section className="flex flex-col gap-1.5 border-b border-border px-4 py-3">
      <h4 className="m-0 flex items-center justify-between text-xs font-semibold text-muted">
        {t("app.heading")}
        <Menu
          trigger={
            <Button size="sm" variant="ghost">
              <IconPlus aria-hidden />
              {t("app.add")}
            </Button>
          }
        >
          {addable.map((app) => (
            <MenuItem
              key={app.id}
              onAction={() => {
                addToChannel.mutate(
                  { appId: app.id, channelId },
                  {
                    onSuccess: () => {
                      toast(t("app.added", { name: app.name }));
                    },
                  },
                );
              }}
            >
              {t("app.addTo", { name: app.name })}
            </MenuItem>
          ))}
          {addable.length > 0 && <MenuSeparator />}
          <MenuItem
            onAction={() => {
              void navigate({ search: openDialog({ dialog: "add-app" }), to: "." });
            }}
          >
            {t("app.create")}
          </MenuItem>
        </Menu>
      </h4>
      {joined.length === 0 && (
        <p className="m-0 text-label font-normal text-muted">{t("app.channelEmpty")}</p>
      )}
      <ul className="m-0 -mx-2 flex list-none flex-col p-0">
        {joined.map((app) => (
          <AppRow
            key={app.id}
            app={app}
            actions={
              <IconButton
                label={t("app.remove", { name: app.name })}
                onPress={() => {
                  removeFromChannel.mutate(
                    { appId: app.id, channelId },
                    {
                      onSuccess: () => {
                        toast(t("app.removed", { name: app.name }));
                      },
                    },
                  );
                }}
              >
                <IconX />
              </IconButton>
            }
          />
        ))}
      </ul>
    </section>
  );
};
