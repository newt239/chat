import { Button, Tooltip } from "@mantine/core";

import type { UserInfo } from "../types";

type ReactionButtonProps = {
  emoji: string;
  users: UserInfo[];
  isActive: boolean;
  onClick: () => void;
};

export const ReactionButton = ({ emoji, users, isActive, onClick }: ReactionButtonProps) => {
  const tooltipLabel = users.map((user) => user.displayName).join("、 ");

  return (
    <Tooltip label={tooltipLabel} position="top" withArrow>
      <Button
        px={8}
        size="xs"
        color="cyan"
        variant={isActive ? "light" : "outline"}
        onClick={onClick}
        radius="full"
      >
        <span className="mr-2">{emoji}</span>
        <span>{users.length}</span>
      </Button>
    </Tooltip>
  );
};
