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
import { useParams, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/ToastRegion/toast";
import {
  useAddBookmark,
  useIsBookmarked,
  useRemoveBookmark,
} from "#/features/bookmark/hooks/useBookmarks";
import { useMentionDirectory } from "#/features/message/hooks/useMentionDirectory";
import { usePinActions } from "#/features/pin/hooks/usePinActions";
import { useToggleThreadFollow } from "#/features/thread/hooks/useToggleThreadFollow";

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

type Options = {
  message: Message;
  threadMetadata: ThreadMetadata | undefined;
  isAuthor: boolean;
  onReplyInThread: () => void;
  onCopyLink: () => void;
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
  onCopyLink,
  onEdit,
  onDelete,
  onViewReactions,
}: Options) => {
  const { t } = useTranslation();
  const { toText } = useMentionDirectory();
  const router = useRouter();
  const { workspaceId } = useParams({ strict: false });
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
    navigator.clipboard.writeText(toText(message.body)).then(
      () => toast(t("message.link.textCopied"), { tone: "success" }),
      () => toast(t("message.link.textCopyFailed"), { tone: "danger" }),
    );
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

  // スレッドのルートができるまでは、親メッセージを指すチャンネルのリンクで代用する
  const threadHref =
    workspaceId === undefined
      ? undefined
      : router.buildLocation({
          params: { channelId: message.channelId, workspaceId },
          search: { message: message.parentId ?? message.id },
          to: "/app/$workspaceId/$channelId",
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
    threadHref !== undefined && {
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
      onAction: onCopyLink,
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
