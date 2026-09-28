import {
  IconBookmark,
  IconBookmarkOff,
  IconExternalLink,
  IconLink,
  IconMessageReply,
  IconPencil,
  IconPin,
  IconPinnedOff,
  IconTrash,
} from "@tabler/icons-react";
import { useParams, useRouter } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { toast } from "#/components/ui/toast";
import {
  useAddBookmark,
  useIsBookmarked,
  useRemoveBookmark,
} from "#/features/bookmark/hooks/useBookmarks";
import { usePinActions } from "#/features/pin/hooks/usePinActions";
import { useIsPinned } from "#/features/pin/hooks/usePinnedMessages";

import type { Message } from "#/gen/chat/v1/message_pb";

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
  isAuthor: boolean;
  onReplyInThread: () => void;
  onCopyLink: () => void;
  onEdit: () => void;
  onDelete: () => void;
};

// ホバー時の「その他」メニューとモバイルのシートで同じ操作を並べる
export const useMessageMenuActions = ({
  message,
  isAuthor,
  onReplyInThread,
  onCopyLink,
  onEdit,
  onDelete,
}: Options) => {
  const { t } = useTranslation();
  const router = useRouter();
  const { workspaceId } = useParams({ strict: false });
  const isBookmarked = useIsBookmarked(message.id);
  const addBookmark = useAddBookmark();
  const removeBookmark = useRemoveBookmark();
  const isPinned = useIsPinned(message.id, message.channelId);
  const { pin, unpin } = usePinActions();
  const canModify = isAuthor && !message.isDeleted;

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
    {
      icon: IconLink,
      id: "copyLink",
      label: t("message.actions.copyLink"),
      onAction: onCopyLink,
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
