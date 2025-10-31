import { useEffect, useState } from "react";

import { Button, Group, Stack, Text, TextInput, Textarea } from "@mantine/core";
import { useAtom } from "jotai";

import { useUpdateProfile } from "@/features/settings/hooks/useUpdateProfile";
import { userAtom } from "@/providers/store/auth";

type Props = {
  onUpdated?: () => void;
};

export const ProfileSettingsPanel = ({ onUpdated }: Props) => {
  const [user] = useAtom(userAtom);
  const mutation = useUpdateProfile();

  const [displayName, setDisplayName] = useState<string>(user?.displayName ?? "");
  const [bio, setBio] = useState<string>("");
  const [avatarUrl, setAvatarUrl] = useState<string>(user?.avatarUrl ?? "");

  useEffect(() => {
    setDisplayName(user?.displayName ?? "");
    setAvatarUrl(user?.avatarUrl ?? "");
  }, [user]);

  const onSubmit = async () => {
    await mutation.mutateAsync({
      displayName: displayName || undefined,
      bio,
      avatarUrl: avatarUrl || null,
    });
    onUpdated?.();
  };

  return (
    <Stack gap="sm">
      <Text fw={600}>プロフィール設定</Text>
      <TextInput label="表示名" value={displayName} onChange={(e) => setDisplayName(e.currentTarget.value)} required />
      <Textarea label="自己紹介" value={bio} onChange={(e) => setBio(e.currentTarget.value)} autosize minRows={3} />
      <TextInput label="アイコンURL" value={avatarUrl} onChange={(e) => setAvatarUrl(e.currentTarget.value)} />
      <Group justify="flex-end">
        <Button onClick={onSubmit} loading={mutation.isPending}>
          保存
        </Button>
      </Group>
      {mutation.isError && (
        <Text c="red" size="sm">{(mutation.error as Error)?.message ?? "更新に失敗しました"}</Text>
      )}
      {mutation.isSuccess && <Text c="green" size="sm">保存しました</Text>}
    </Stack>
  );
};


