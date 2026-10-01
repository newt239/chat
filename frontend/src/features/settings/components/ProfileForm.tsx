import { useState } from "react";

import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconImageField } from "#/components/block/IconImageField/IconImageField";
import { Button } from "#/components/ui/Button/Button";
import { TextArea } from "#/components/ui/TextArea/TextArea";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { ImagePurpose } from "#/gen/chat/v1/image_service_pb";

import { useUpdateProfile } from "../hooks/useUpdateProfile";
import { ProfileLinksField } from "./ProfileLinksField";

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
  const [links, setLinks] = useState(me.links);
  // 空の行は保存しない
  const filledLinks = links.map((link) => link.trim()).filter((link) => link !== "");
  const isDirty =
    displayName !== me.displayName ||
    bio !== (me.bio ?? "") ||
    avatarUrl !== (me.avatarUrl ?? "") ||
    filledLinks.join("\n") !== me.links.join("\n");

  return (
    <Form
      className="flex flex-col gap-3.5 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        mutation.mutate(
          { avatarUrl, bio, displayName, links: { urls: filledLinks } },
          {
            onSuccess: () => {
              toast(t("settings.profile.saved"), { tone: "success" });
            },
          },
        );
      }}
    >
      <div className="flex min-w-0 flex-col">
        <b className="truncate text-body-strong">{displayName || me.displayName}</b>
        <span className="truncate text-caption text-muted">{me.email}</span>
      </div>
      <IconImageField
        label={t("settings.profile.avatar")}
        name={displayName || me.displayName}
        value={avatarUrl}
        onChange={setAvatarUrl}
        purpose={ImagePurpose.AVATAR}
        workspaceId={null}
      />
      <TextField
        label={t("auth.displayName")}
        description={t("settings.profile.displayNameDescription")}
        value={displayName}
        onChange={setDisplayName}
        isRequired
      />
      <TextArea label={t("settings.profile.bio")} value={bio} onChange={setBio} />
      <ProfileLinksField value={links} onChange={setLinks} />
      {mutation.isError && (
        <p role="alert" className="m-0 text-caption text-danger">
          {mutation.error.message}
        </p>
      )}
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
