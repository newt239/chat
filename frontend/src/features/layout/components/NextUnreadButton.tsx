import { IconChevronsDown } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { useNextUnreadChannel } from "#/features/channel/hooks/useNextUnreadChannel";

type NextUnreadButtonProps = {
  workspaceId: string;
};

export const NextUnreadButton = ({ workspaceId }: NextUnreadButtonProps) => {
  const { t } = useTranslation();
  const { goNext, nextId } = useNextUnreadChannel(workspaceId);
  if (nextId === undefined) {
    return null;
  }
  return (
    <Tooltip content={t("shell.channel.nextUnread")}>
      <Button onPress={goNext} className="rounded-full px-3.5">
        <IconChevronsDown aria-hidden />
        {t("shell.channel.nextUnreadShort")}
      </Button>
    </Tooltip>
  );
};
