import { Modal, Button, Text, Group, Stack } from "@mantine/core";
import { IconLogout } from "@tabler/icons-react";
import { useSetAtom } from "jotai";

import { clearAuthAtom } from "@/providers/store/auth";

type SettingsModalProps = {
  opened: boolean;
  onClose: () => void;
};

export const SettingsModal = ({ opened, onClose }: SettingsModalProps) => {
  const clearAuth = useSetAtom(clearAuthAtom);

  const handleLogout = () => {
    clearAuth();
    onClose();
  };

  return (
    <Modal opened={opened} onClose={onClose} title="設定" centered size="sm">
      <Stack gap="md">
        <Text size="sm" c="dimmed">
          アプリケーションの設定とアカウント管理を行えます。
        </Text>

        <Group justify="flex-end">
          <Button variant="outline" onClick={onClose}>
            キャンセル
          </Button>
          <Button
            variant="filled"
            color="red"
            leftSection={<IconLogout size={16} />}
            onClick={handleLogout}
          >
            ログアウト
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
