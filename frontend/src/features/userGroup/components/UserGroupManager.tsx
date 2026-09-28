import { useState } from "react";

import { IconTrash } from "@tabler/icons-react";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { IconButton } from "#/components/ui/IconButton";
import { TextField } from "#/components/ui/TextField";
import { UserGroupMembers } from "#/features/userGroup/components/UserGroupMembers";
import { useUserGroupActions, useUserGroups } from "#/features/userGroup/hooks/useUserGroups";

type UserGroupManagerProps = {
  workspaceId: string;
};

export const UserGroupManager = ({ workspaceId }: UserGroupManagerProps) => {
  const { t } = useTranslation();
  const { data: groups } = useUserGroups(workspaceId);
  const { create, remove } = useUserGroupActions();
  const [name, setName] = useState("");

  return (
    <div className="flex flex-col gap-3 bg-surface font-sans text-text">
      <h3 className="m-0 text-[13px] font-bold">
        {t("userGroup.title", { count: groups?.length ?? 0 })}
      </h3>

      {groups?.length === 0 && (
        <p className="m-0 text-caption text-muted">{t("userGroup.empty")}</p>
      )}

      {groups?.map((group) => (
        <article
          key={group.id}
          className="flex flex-col gap-2 rounded-xl border border-border bg-surface p-3.5"
        >
          <header className="flex items-start gap-2">
            <div className="flex min-w-0 flex-1 flex-col gap-0.5">
              <span className="text-sm font-bold text-accent-text">@{group.name}</span>
              {group.description !== undefined && group.description.length > 0 && (
                <p className="m-0 text-[12.5px] text-muted">{group.description}</p>
              )}
            </div>
            <IconButton
              label={t("userGroup.delete", { name: group.name })}
              onPress={() => {
                remove.mutate({ groupId: group.id });
              }}
            >
              <IconTrash />
            </IconButton>
          </header>
          <UserGroupMembers groupId={group.id} workspaceId={workspaceId} />
        </article>
      ))}

      <Form
        className="flex items-end gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          create.mutate({ name: name.trim(), workspaceId });
          setName("");
        }}
      >
        <TextField
          label={t("userGroup.createLabel")}
          placeholder={t("userGroup.createPlaceholder")}
          value={name}
          onChange={setName}
          className="flex-1"
        />
        <Button type="submit" isDisabled={name.trim().length === 0} isPending={create.isPending}>
          {t("userGroup.create")}
        </Button>
      </Form>

      {create.isError && <p className="m-0 text-caption text-danger">{create.error.message}</p>}
    </div>
  );
};
