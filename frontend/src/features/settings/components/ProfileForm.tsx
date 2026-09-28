import { useState } from "react";

import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Button } from "#/components/ui/Button";
import { TextArea } from "#/components/ui/TextArea";
import { TextField } from "#/components/ui/TextField";
import { toast } from "#/components/ui/toast";

import { useUpdateProfile } from "../hooks/useUpdateProfile";

import type { User } from "#/gen/chat/v1/user_pb";

type ProfileFormProps = {
  me: User;
};

export const ProfileForm = ({ me }: ProfileFormProps) => {
  const { t } = useTranslation();
  const mutation = useUpdateProfile();
  const [displayName, setDisplayName] = useState(me.displayName);
  const [bio, setBio] = useState(me.bio ?? "");
  const [avatarUrl, setAvatarUrl] = useState(me.avatarUrl ?? "");
  const isDirty =
    displayName !== me.displayName || bio !== (me.bio ?? "") || avatarUrl !== (me.avatarUrl ?? "");

  return (
    <Form
      className="flex flex-col gap-3.5 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        mutation.mutate(
          { avatarUrl: avatarUrl || undefined, bio, displayName },
          {
            onSuccess: () => {
              toast(t("settings.profile.saved"), { tone: "success" });
            },
          },
        );
      }}
    >
      <div className="flex items-center gap-3">
        <Avatar name={displayName || me.displayName} src={avatarUrl || null} size={64} />
        <div className="flex min-w-0 flex-col">
          <b className="truncate text-body-strong">{displayName || me.displayName}</b>
          <span className="truncate text-caption text-muted">{me.email}</span>
        </div>
      </div>
      <TextField
        label={t("auth.displayName")}
        description={t("settings.profile.displayNameDescription")}
        value={displayName}
        onChange={setDisplayName}
        isRequired
      />
      <TextArea label={t("settings.profile.bio")} value={bio} onChange={setBio} />
      <TextField
        label={t("settings.profile.avatarUrl")}
        type="url"
        placeholder="https://"
        value={avatarUrl}
        onChange={setAvatarUrl}
        errorMessage={mutation.isError ? mutation.error.message : undefined}
      />
      <Button
        type="submit"
        className="self-start"
        isDisabled={!isDirty}
        isPending={mutation.isPending}
      >
        {t("common.save")}
      </Button>
    </Form>
  );
};
