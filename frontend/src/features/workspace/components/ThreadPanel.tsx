import { Stack, Text } from "@mantine/core";

const SIDEBAR_CONTAINER_CLASS = "border-l border-gray-200 bg-gray-50 p-4 h-full overflow-y-auto";

type ThreadPanelProps = {
  threadId: string;
};

export const ThreadPanel = ({ threadId }: ThreadPanelProps) => {
  return (
    <div className={SIDEBAR_CONTAINER_CLASS}>
      <Stack gap="md">
        <div>
          <Text size="sm" fw={600}>
            スレッド
          </Text>
          <Text size="xs" c="dimmed">
            ID: {threadId}
          </Text>
        </div>
        <Text size="sm" c="dimmed">
          スレッドの詳細表示は現在準備中です。
        </Text>
      </Stack>
    </div>
  );
};
