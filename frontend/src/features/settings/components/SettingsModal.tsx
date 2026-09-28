import { Modal, Button, Text, Group, Stack, Divider } from "@mantine/core";
import { IconLogout } from "@tabler/icons-react";

import { useLogout } from "#/features/auth/hooks/useLogout";
import { AccountSettingsPanel } from "#/features/settings/components/AccountSettingsPanel";
import { ProfileSettingsPanel } from "#/features/settings/components/ProfileSettingsPanel";

type SettingsModalProps = {
  opened: boolean;
  onClose: () => void;
};

export const SettingsModal = ({ opened, onClose }: SettingsModalProps) => {
  const logout = useLogout();

  const handleLogout = () => {
    onClose();
    logout.mutate({});
  };

  return (
    <Modal opened={opened} onClose={onClose} title="設定" centered size="md">
      <Stack gap="md">
        <Text size="sm" c="dimmed">
          アプリケーションの設定とアカウント管理を行えます。
        </Text>

        <ProfileSettingsPanel />

        <Divider />

        <AccountSettingsPanel />

        <Divider />

        <Group justify="flex-end">
          <Button variant="outline" onClick={onClose}>
            キャンセル
          </Button>
          <Button
            variant="filled"
            color="red"
            leftSection={<IconLogout size={16} />}
            loading={logout.isPending}
            onClick={handleLogout}
          >
            ログアウト
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
