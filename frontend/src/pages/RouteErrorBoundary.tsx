import { Button, Stack, Text } from "@mantine/core";
import { useRouteError } from "react-router";

import { logger } from "#/lib/logger";
import { paths } from "#/lib/paths";

export const RouteErrorBoundary = () => {
  const error = useRouteError();
  logger.error("ルーティングエラー:", error);

  return (
    <Stack align="center" justify="center" className="h-full" gap="md">
      <Text size="lg" fw={600}>
        問題が発生しました
      </Text>
      <Text size="sm" c="dimmed">
        ページの読み込み中にエラーが発生しました。
      </Text>
      <Button component="a" href={paths.app()}>
        トップへ戻る
      </Button>
    </Stack>
  );
};
