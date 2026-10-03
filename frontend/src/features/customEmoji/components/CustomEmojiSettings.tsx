import { useState } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { IconMoodPlus, IconTrash } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { TextField } from "#/components/ui/TextField/TextField";
import { Permission, PermissionService } from "#/gen/chat/v1/permission_service_pb";

import { useCustomEmojiActions } from "../hooks/useCustomEmojiActions";
import { useCustomEmojis } from "../hooks/useCustomEmojis";
import { toCustomEmojiValue } from "../utils/customEmoji";
import { CustomEmojiForm } from "./CustomEmojiForm";

import type { CustomEmoji } from "#/gen/chat/v1/custom_emoji_service_pb";

type CustomEmojiSettingsProps = {
  workspaceId: string;
};

export const CustomEmojiSettings = ({ workspaceId }: CustomEmojiSettingsProps) => {
  const { t } = useTranslation();
  const { data: emojis } = useCustomEmojis(workspaceId);
  const { data: canCreate = false } = useQuery(
    PermissionService.method.getPermissions,
    { workspaceId },
    { select: (res) => res.myPermissions.includes(Permission.CREATE_CUSTOM_EMOJI) },
  );
  const [query, setQuery] = useState("");
  const [deleting, setDeleting] = useState<CustomEmoji | null>(null);
  const { remove } = useCustomEmojiActions(workspaceId);

  const keyword = query.trim().toLowerCase().replaceAll(":", "");
  const filtered = emojis?.filter((emoji) => emoji.name.includes(keyword));

  return (
    <div className="flex flex-col gap-5">
      <p className="m-0 text-caption text-muted">{t("workspace.emoji.description")}</p>
      {canCreate && (
        <section className="flex flex-col gap-3 rounded-lg border border-border p-3">
          <h3 className="m-0 text-body-strong">{t("workspace.emoji.add")}</h3>
          <CustomEmojiForm workspaceId={workspaceId} onAdded={() => {}} />
        </section>
      )}
      <section className="flex flex-col gap-2">
        <TextField
          label={t("workspace.emoji.search")}
          type="search"
          value={query}
          onChange={setQuery}
          placeholder=":party:"
        />
        {filtered === undefined ? (
          <Skeleton className="h-40 w-full rounded-xl" />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={<IconMoodPlus />}
            title={t(query === "" ? "workspace.emoji.empty" : "workspace.emoji.noMatch")}
            description={t("workspace.emoji.emptyDescription")}
          />
        ) : (
          <ul className="m-0 flex list-none flex-col p-0">
            {filtered.map((emoji) => (
              <li
                key={emoji.id}
                className="flex items-center gap-3 border-b border-border py-2 last:border-b-0"
              >
                <img
                  src={emoji.imageUrl}
                  alt=""
                  loading="lazy"
                  className="size-8 shrink-0 object-contain"
                />
                <span className="flex min-w-0 flex-1 flex-col">
                  <span className="truncate font-mono text-body">
                    {toCustomEmojiValue(emoji.name)}
                  </span>
                  <span className="truncate text-caption text-muted">
                    {t("workspace.emoji.createdBy", {
                      name: emoji.createdBy?.displayName ?? "",
                    })}
                  </span>
                </span>
                {emoji.canDelete && (
                  <IconButton
                    label={t("workspace.emoji.delete", { name: toCustomEmojiValue(emoji.name) })}
                    onPress={() => {
                      setDeleting(emoji);
                    }}
                  >
                    <IconTrash />
                  </IconButton>
                )}
              </li>
            ))}
          </ul>
        )}
      </section>

      <AlertDialog
        isOpen={deleting !== null}
        onOpenChange={() => {
          setDeleting(null);
        }}
        title={t("workspace.emoji.deleteConfirm", {
          name: deleting === null ? "" : toCustomEmojiValue(deleting.name),
        })}
        confirmLabel={t("common.delete")}
        tone="danger"
        isPending={remove.isPending}
        onConfirm={() => {
          if (deleting === null) {
            return;
          }
          remove.mutate(
            { emojiId: deleting.id, workspaceId },
            {
              onSuccess: () => {
                setDeleting(null);
              },
            },
          );
        }}
      >
        {t("workspace.emoji.deleteDescription")}
      </AlertDialog>
    </div>
  );
};
