import { IconExternalLink } from "@tabler/icons-react";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
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

import { closePanel } from "../utils/overlaySearch";
import { workspaceRoute } from "../utils/workspaceRoute";

/** 右パネル（モバイルでは全画面のページ）に出す内容と閉じる操作。 ?profile= などの search がスレッドのパスより優先する */
export const useRightPanel = (workspaceId: string) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { channelId, messageId } = useParams({ strict: false });
  const { group, panel, profile } = workspaceRoute.useSearch();
  const myId = useAtomValue(userAtom)?.id;
  const channelPanel = channelId === undefined ? undefined : panel;
  const isThreadOpen = messageId !== undefined && channelId !== undefined;

  const close = () => {
    if (
      profile === undefined &&
      group === undefined &&
      channelPanel === undefined &&
      isThreadOpen
    ) {
      void navigate({ params: { channelId, workspaceId }, to: "/app/$workspaceId/$channelId" });
      return;
    }
    void navigate({ search: closePanel, to: "." });
  };

  const content = (() => {
    if (profile !== undefined) {
      // 自分のプロフィールはこのパネルで編集する
      return profile === myId
        ? {
            body: <ProfileEditor />,
            extra: null,
            key: "profile-me",
            title: t("shell.rightPanel.myProfile"),
          }
        : {
            body: <UserProfilePanel workspaceId={workspaceId} userId={profile} />,
            extra: null,
            key: `profile-${profile}`,
            title: t("shell.rightPanel.profile"),
          };
    }
    if (group !== undefined) {
      return {
        body: <UserGroupPanel workspaceId={workspaceId} groupId={group} />,
        extra: null,
        key: `group-${group}`,
        title: t("shell.rightPanel.userGroup"),
      };
    }
    if (channelId !== undefined && channelPanel !== undefined) {
      return {
        info: {
          body: <ChannelInfoPanel workspaceId={workspaceId} channelId={channelId} />,
          extra: null,
          key: `info-${channelId}`,
          title: t("shell.rightPanel.channelInfo"),
        },
        members: {
          body: <ChannelMemberPanel channelId={channelId} />,
          extra: null,
          key: `members-${channelId}`,
          title: t("shell.rightPanel.members"),
        },
        pins: {
          body: <PinnedPanel channelId={channelId} />,
          extra: null,
          key: `pins-${channelId}`,
          title: t("shell.rightPanel.pins"),
        },
      }[channelPanel];
    }
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
    return null;
  })();

  return { close, content };
};
