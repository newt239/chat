import { useEffect, useState } from "react";

import { Button, Group, Stack, Text, TextInput, Textarea } from "@mantine/core";

import { useMe } from "#/features/settings/hooks/useMe";
import { useUpdateProfile } from "#/features/settings/hooks/useUpdateProfile";

type Props = {
  onUpdated?: () => void;
};

export const ProfileSettingsPanel = ({ onUpdated }: Props) => {
  const { data: me } = useMe();
  const mutation = useUpdateProfile();

  const [displayName, setDisplayName] = useState("");
  const [bio, setBio] = useState("");
  const [avatarUrl, setAvatarUrl] = useState("");

  useEffect(() => {
    setDisplayName(me?.displayName ?? "");
    setBio(me?.bio ?? "");
    setAvatarUrl(me?.avatarUrl ?? "");
  }, [me]);

  const onSubmit = async () => {
    await mutation.mutateAsync({
      avatarUrl: avatarUrl || undefined,
      bio,
      displayName: displayName || undefined,
    });
    onUpdated?.();
  };

  return (
    <Stack gap="sm">
      <Text fw={600}>プロフィール設定</Text>
      <TextInput
        label="表示名"
        value={displayName}
        onChange={(e) => {
          setDisplayName(e.currentTarget.value);
        }}
        required
      />
      <Textarea
        label="自己紹介"
        value={bio}
        onChange={(e) => {
          setBio(e.currentTarget.value);
        }}
        autosize
        minRows={3}
      />
      <TextInput
        label="アイコンURL"
        value={avatarUrl}
        onChange={(e) => {
          setAvatarUrl(e.currentTarget.value);
        }}
      />
      <Group justify="flex-end">
        <Button
          onClick={() => {
            void onSubmit();
          }}
          loading={mutation.isPending}
        >
          保存
        </Button>
      </Group>
      {mutation.isError && (
        <Text c="red" size="sm">
          {mutation.error.message}
        </Text>
      )}
      {mutation.isSuccess && (
        <Text c="green" size="sm">
          保存しました
        </Text>
      )}
    </Stack>
  );
};
