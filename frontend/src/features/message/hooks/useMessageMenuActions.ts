import {
  IconBell,
  IconBellOff,
  IconBookmark,
  IconBookmarkOff,
  IconCopy,
  IconExternalLink,
  IconLink,
  IconList,
  IconMessageReply,
  IconPencil,
  IconPin,
  IconPinnedOff,
  IconTrash,
} from "@tabler/icons-react";
import { useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import {
  useAddBookmark,
  useIsBookmarked,
  useRemoveBookmark,
} from "#/features/bookmark/hooks/useBookmarks";
import { workspaceRoute } from "#/features/layout/utils/workspaceRoute";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { usePinActions } from "#/features/pin/hooks/usePinActions";
import { useToggleThreadFollow } from "#/features/thread/hooks/useToggleThreadFollow";
import { copyWithToast } from "#/lib/clipboard";
import { toShareUrl } from "#/lib/shareUrl";

import { messageLocation } from "../utils/messageLocation";

import type { Message, ThreadMetadata } from "#/gen/chat/v1/message_pb";

import type { Icon } from "@tabler/icons-react";

export type MessageMenuAction = {
  id: string;
  label: string;
  icon: Icon;
  tone: "default" | "danger";
  // 新しいタブで開くリンク。ブラウザ標準の動作に任せるため onAction は持たない
  href?: string;
  onAction?: () => void;
};

// ツールバーは先頭の 3 つ、モバイルのシートはすべてを並べる
export const quickReactions = ["👍", "✅", "👀", "🎉", "🙏"] as const;

type Options = {
  message: Message;
  threadMetadata: ThreadMetadata | undefined;
  isAuthor: boolean;
  onReplyInThread: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onViewReactions: () => void;
};

// ホバー時の「その他」メニューとモバイルのシートで同じ操作を並べる
export const useMessageMenuActions = ({
  message,
  threadMetadata,
  isAuthor,
  onReplyInThread,
  onEdit,
  onDelete,
  onViewReactions,
}: Options) => {
  const { t } = useTranslation();
  const { toText } = useMentionDirectory();
  const router = useRouter();
  const { workspaceId } = workspaceRoute.useParams();
  const isBookmarked = useIsBookmarked(message.id);
  const addBookmark = useAddBookmark();
  const removeBookmark = useRemoveBookmark();
  const isPinned = message.pin !== undefined;
  const { pin, unpin } = usePinActions();
  const canModify = isAuthor && !message.isDeleted;
  const { setFollowing } = useToggleThreadFollow(message.id);
  const isFollowing = threadMetadata?.isFollowing ?? false;

  const toggleBookmark = () => {
    (isBookmarked ? removeBookmark : addBookmark).mutate(
      { messageId: message.id },
      {
        onSuccess: () => {
          toast(t(isBookmarked ? "bookmark.removed" : "bookmark.added"));
        },
      },
    );
  };

  const copyText = () => {
    void copyWithToast(toText(message.body), t("message.link.textCopied"));
  };

  const togglePin = () => {
    (isPinned ? unpin : pin).mutate(
      { channelId: message.channelId, messageId: message.id },
      {
        onSuccess: () => {
          toast(t(isPinned ? "pin.unpinned" : "pin.pinned"));
        },
      },
    );
  };

  const copyLink = () => {
    const { href } = router.buildLocation(
      messageLocation({
        channelId: message.channelId,
        messageId: message.id,
        parentId: message.parentId,
        workspaceId,
      }),
    );
    void copyWithToast(toShareUrl(href), t("message.link.copied"));
  };

  const threadHref = router.buildLocation({
    params: {
      channelId: message.channelId,
      messageId: message.parentId ?? message.id,
      workspaceId,
    },
    search: { message: message.parentId === undefined ? undefined : message.id },
    to: "/app/$workspaceId/$channelId/thread/$messageId",
  }).href;

  const actions: (MessageMenuAction | false)[] = [
    canModify && {
      icon: IconPencil,
      id: "edit",
      label: t("message.actions.edit"),
      onAction: onEdit,
      tone: "default",
    },
    {
      icon: IconMessageReply,
      id: "thread",
      label: t("message.actions.replyInThread"),
      onAction: onReplyInThread,
      tone: "default",
    },
    (threadMetadata?.replyCount ?? 0) > 0 && {
      icon: isFollowing ? IconBellOff : IconBell,
      id: "followThread",
      label: t(isFollowing ? "thread.follow.unfollow" : "thread.follow.follow"),
      onAction: () => {
        setFollowing(!isFollowing);
      },
      tone: "default",
    },
    {
      href: threadHref,
      icon: IconExternalLink,
      id: "threadInNewTab",
      label: t("message.actions.openThreadInNewTab"),
      tone: "default",
    },
    {
      icon: isBookmarked ? IconBookmarkOff : IconBookmark,
      id: "bookmark",
      label: t(isBookmarked ? "message.actions.unbookmark" : "message.actions.bookmark"),
      onAction: toggleBookmark,
      tone: "default",
    },
    {
      icon: isPinned ? IconPinnedOff : IconPin,
      id: "pin",
      label: t(isPinned ? "message.actions.unpin" : "message.actions.pin"),
      onAction: togglePin,
      tone: "default",
    },
    message.reactions.length > 0 && {
      icon: IconList,
      id: "reactions",
      label: t("reaction.list.open"),
      onAction: onViewReactions,
      tone: "default",
    },
    {
      icon: IconLink,
      id: "copyLink",
      label: t("message.actions.copyLink"),
      onAction: copyLink,
      tone: "default",
    },
    !message.isDeleted &&
      message.body.trim() !== "" && {
        icon: IconCopy,
        id: "copyText",
        label: t("message.actions.copyText"),
        onAction: copyText,
        tone: "default",
      },
    canModify && {
      icon: IconTrash,
      id: "delete",
      label: t("message.actions.delete"),
      onAction: onDelete,
      tone: "danger",
    },
  ];

  return {
    actions: actions.filter((action) => action !== false),
    isBookmarked,
    toggleBookmark,
  };
};
