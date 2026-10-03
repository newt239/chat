import { IconBell } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconToggleButton } from "#/components/ui/IconToggleButton/IconToggleButton";

import { useToggleThreadFollow } from "../hooks/useToggleThreadFollow";

type ThreadFollowButtonProps = {
  threadId: string;
  isFollowing: boolean;
  className?: string;
};

export const ThreadFollowButton = ({
  threadId,
  isFollowing,
  className,
}: ThreadFollowButtonProps) => {
  const { t } = useTranslation();
  const { isPending, setFollowing } = useToggleThreadFollow(threadId);

  return (
    <IconToggleButton
      label={t("thread.follow.follow")}
      className={className}
      isSelected={isFollowing}
      isDisabled={isPending}
      onChange={setFollowing}
    >
      <IconBell aria-hidden />
    </IconToggleButton>
  );
};
