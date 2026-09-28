import { Button, Stack, Text } from "@mantine/core";
import { Link } from "@tanstack/react-router";

import { logger } from "#/lib/logger";

import type { ErrorComponentProps } from "@tanstack/react-router";

export const RouteErrorBoundary = ({ error }: ErrorComponentProps) => {
  logger.error("ルーティングエラー:", error);

  return (
    <Stack align="center" justify="center" className="h-full" gap="md">
      <Text size="lg" fw={600}>
        問題が発生しました
      </Text>
      <Text size="sm" c="dimmed">
        ページの読み込み中にエラーが発生しました。
      </Text>
      <Button renderRoot={(props) => <Link {...props} to="/app" reloadDocument />}>
        トップへ戻る
      </Button>
    </Stack>
  );
};
