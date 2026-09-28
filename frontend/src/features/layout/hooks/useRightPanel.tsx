import { IconExternalLink } from "@tabler/icons-react";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";
import { useTranslation } from "react-i18next";

import { LinkButton } from "#/components/ui/LinkButton";
import { Tooltip } from "#/components/ui/Tooltip";
import { ChannelInfoPanel } from "#/features/channel/components/ChannelInfoPanel";
import { ChannelMemberPanel } from "#/features/channel/components/ChannelMemberPanel";
import { UserProfilePanel } from "#/features/member/components/UserProfilePanel";
import { PinnedPanel } from "#/features/pin/components/PinnedPanel";
import { ProfileEditor } from "#/features/settings/components/ProfileEditor";
import { ThreadPanel } from "#/features/thread/components/ThreadPanel";
import { UserGroupPanel } from "#/features/userGroup/components/UserGroupPanel";
import { userAtom } from "#/providers/store/auth";
import { closeRightSidePanelAtom, rightSidePanelViewAtom } from "#/providers/store/ui";

/** 右パネル（モバイルでは全画面のページ）に出す内容と閉じる操作。 スレッドは URL、それ以外は rightSidePanelViewAtom で決まる */
export const useRightPanel = (workspaceId: string) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { channelId, messageId } = useParams({ strict: false });
  const view = useAtomValue(rightSidePanelViewAtom);
  const closeRightPanel = useSetAtom(closeRightSidePanelAtom);
  const myId = useAtomValue(userAtom)?.id;
  const isThreadOpen = messageId !== undefined && channelId !== undefined;

  const close = () => {
    if (isThreadOpen) {
      void navigate({ params: { channelId, workspaceId }, to: "/app/$workspaceId/$channelId" });
      return;
    }
    closeRightPanel();
  };

  const content = (() => {
    if (isThreadOpen) {
      return {
        body: <ThreadPanel workspaceId={workspaceId} channelId={channelId} threadId={messageId} />,
        extra: (
          <Tooltip content={t("shell.openInNewTab")}>
            <LinkButton
              variant="ghost"
              aria-label={t("shell.openInNewTab")}
              className="size-[30px] px-0 [&_svg]:size-4"
              to="/app/$workspaceId/$channelId/thread/$messageId"
              params={{ channelId, messageId, workspaceId }}
              target="_blank"
            >
              <IconExternalLink aria-hidden />
            </LinkButton>
          </Tooltip>
        ),
        key: `thread-${messageId}`,
        title: t("shell.rightPanel.thread"),
      };
    }
    switch (view.type) {
      case "channel-members": {
        return {
          body: <ChannelMemberPanel channelId={view.channelId} />,
          extra: null,
          key: `members-${view.channelId}`,
          title: t("shell.rightPanel.members"),
        };
      }
      case "channel-info": {
        return {
          body: <ChannelInfoPanel workspaceId={workspaceId} channelId={view.channelId} />,
          extra: null,
          key: `info-${view.channelId ?? ""}`,
          title: t("shell.rightPanel.channelInfo"),
        };
      }
      case "pins": {
        return {
          body: <PinnedPanel channelId={view.channelId} />,
          extra: null,
          key: `pins-${view.channelId}`,
          title: t("shell.rightPanel.pins"),
        };
      }
      case "user-profile": {
        // 自分のプロフィールはこのパネルで編集する
        return view.userId === myId
          ? {
              body: <ProfileEditor />,
              extra: null,
              key: "profile-me",
              title: t("shell.rightPanel.myProfile"),
            }
          : {
              body: <UserProfilePanel workspaceId={workspaceId} userId={view.userId} />,
              extra: null,
              key: `profile-${view.userId}`,
              title: t("shell.rightPanel.profile"),
            };
      }
      case "user-group": {
        return {
          body: <UserGroupPanel workspaceId={workspaceId} groupId={view.groupId} />,
          extra: null,
          key: `group-${view.groupId}`,
          title: t("shell.rightPanel.userGroup"),
        };
      }
      case "hidden": {
        return null;
      }
      default: {
        return null;
      }
    }
  })();

  return { close, content };
};
