import { Button, Tooltip } from "@mantine/core";

import type { UserInfo } from "../types";

type ReactionButtonProps = {
  emoji: string;
  count: number;
  users: UserInfo[];
  isActive: boolean;
  onClick: () => void;
}

export const ReactionButton = ({
  emoji,
  count,
  users,
  isActive,
  onClick,
}: ReactionButtonProps) => {
  const tooltipLabel = users.map((user) => user.displayName).join("、 ");

  return (
    <Tooltip label={tooltipLabel} position="top" withArrow>
      <Button
        size="xs"
        variant={isActive ? "filled" : "light"}
        onClick={onClick}
        styles={{
          root: {
            padding: "4px 8px",
            height: "auto",
            fontSize: "14px",
          },
        }}
      >
        <span style={{ marginRight: "4px" }}>{emoji}</span>
        <span>{count}</span>
      </Button>
    </Tooltip>
  );
};
