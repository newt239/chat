import { useState } from "react";
import type { ReactNode } from "react";

import { IconBookmarkFilled, IconPin } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Badge } from "#/components/ui/Badge/Badge";
import { cn, focusRing } from "#/components/ui/styles/styles";
import { MessageAttachments } from "#/features/attachment/components/MessageAttachments";
import { closeDialog, openDialog, openPanel } from "#/features/layout/utils/overlaySearch";
import { workspaceRoute } from "#/features/layout/utils/workspaceRoute";
import { MessageLocationCard } from "#/features/location/components/MessageLocationCard";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { MessagePollCard } from "#/features/poll/components/MessagePollCard";
import { ReactionList } from "#/features/reaction/components/ReactionList";
import { ReactionsDialog } from "#/features/reaction/components/ReactionsDialog";
import { useToggleReaction } from "#/features/reaction/hooks/useReactions";
import { ALL_REACTIONS_TAB } from "#/features/reaction/utils/reactionTabs";
import { useIsMobile } from "#/hooks/useMediaQuery";
import { toDate } from "#/lib/timestamp";
import { myUserIdAtom } from "#/providers/store/auth";

import { useLongPress } from "../hooks/useLongPress";
import { useMessageActions } from "../hooks/useMessageActions";
import { useMessageMenuActions } from "../hooks/useMessageMenuActions";
import { useOwnsMessageOverlay } from "../hooks/useOwnsMessageOverlay";
import { MessageActionSheet } from "./MessageActionSheet";
import { MessageContent } from "./MessageContent";
import { MessageEditor } from "./MessageEditor";
import { MessageTime } from "./MessageTime";
import { MessageToolbar } from "./MessageToolbar";
import { ThreadMetadataPreview } from "./ThreadMetadataPreview";

import type { Message, ThreadMetadata } from "#/gen/chat/v1/message_pb";

type MessageItemProps = {
  message: Message;
  onCopyLink: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  threadMetadata?: ThreadMetadata;
  isHighlighted?: boolean;
  // 親チャンネルの集約表示で、子孫チャンネルのメッセージに付けるチップ
  channelChip?: ReactNode;
};

export const MessageItem = ({
  message,
  onCopyLink,
  onCreateThread,
  threadMetadata,
  isHighlighted = false,
  channelChip = null,
}: MessageItemProps) => {
  const { t } = useTranslation();
  const myId = useAtomValue(myUserIdAtom);
  const navigate = useNavigate();
  const isMobile = useIsMobile();
  const ownsOverlay = useOwnsMessageOverlay(message.id);
  // リアクション一覧（?reactions=&emoji=）と操作シート（?sheet=）は URL で開く
  const reactionTab = workspaceRoute.useSearch({
    select: (search) =>
      ownsOverlay && search.reactions === message.id ? (search.emoji ?? ALL_REACTIONS_TAB) : null,
  });
  const isSheetOpen = workspaceRoute.useSearch({
    select: (search) => ownsOverlay && search.sheet === message.id,
  });
  const [isEditing, setIsEditing] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isOverlayOpen, setIsOverlayOpen] = useState(false);
  const [isFocusWithin, setIsFocusWithin] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const setReactionTab = (tab: string | null) => {
    void navigate({
      // タブの切り替えは履歴に積まない
      replace: reactionTab !== null && tab !== null,
      search:
        tab === null
          ? closeDialog
          : openDialog({
              emoji: tab === ALL_REACTIONS_TAB ? undefined : tab,
              reactions: message.id,
            }),
      to: ".",
    });
  };
  const { isPressed, longPressProps } = useLongPress(() => {
    void navigate({ search: openDialog({ sheet: message.id }), to: "." });
  }, isMobile && !isEditing);
  const { handleEdit, handleDelete, isDeleting } = useMessageActions();
  const toggleReaction = useToggleReaction(message.id);
  // グループ経由は投稿時点のメンバーに展開済み。@channel / @here はチャンネルのメンバー全員宛て
  const isMentioned =
    message.mentionsChannel ||
    message.mentionsHere ||
    message.mentions.some((mention) => mention.userId === myId);

  const { actions, isBookmarked, toggleBookmark } = useMessageMenuActions({
    isAuthor: message.userId === myId,
    message,
    onCopyLink: () => {
      onCopyLink(message.id);
    },
    onDelete: () => {
      setIsDeleteOpen(true);
    },
    onEdit: () => {
      setIsEditing(true);
    },
    onReplyInThread: () => {
      onCreateThread(message.id);
    },
    onViewReactions: () => {
      setReactionTab(ALL_REACTIONS_TAB);
    },
    threadMetadata,
  });

  const react = (emoji: string) => {
    toggleReaction(
      emoji,
      message.reactions.some((reaction) => reaction.emoji === emoji && reaction.user?.id === myId),
    );
  };

  const openProfile = () => {
    void navigate({ search: openPanel({ profile: message.userId }), to: "." });
  };

  const displayName = useDisplayName()(message.userId, message.user?.displayName ?? "");
  // アプリの投稿者はプロフィールを持たないため開かない
  const isApp = message.user?.isApp ?? false;
  const avatar = (
    <Avatar name={displayName} src={message.user?.avatarUrl} size={isMobile ? 34 : 32} />
  );
  const createdAt = toDate(message.createdAt);
  const showToolbar =
    !isMobile && !isEditing && !message.isDeleted && (isHovered || isFocusWithin || isOverlayOpen);

  return (
    <div
      onPointerEnter={(event) => {
        // タッチでは疑似的なホバーでツールバーを出さない
        setIsHovered(event.pointerType === "mouse");
      }}
      onPointerLeave={() => {
        setIsHovered(false);
      }}
      onFocus={() => {
        setIsFocusWithin(true);
      }}
      onBlur={(event) => {
        setIsFocusWithin(event.currentTarget.contains(event.relatedTarget));
      }}
      // モバイルでは長押しの onPointerLeave を優先する
      {...longPressProps}
      data-message-id={message.id}
      className={cn(
        "relative flex gap-2.5 px-[18px] py-1.5 font-sans text-text",
        (isHovered || isOverlayOpen) && "bg-hover",
        isMobile && "select-none [-webkit-touch-callout:none]",
        isPressed && "bg-hover",
        isHighlighted
          ? "bg-accent-soft"
          : isMentioned
            ? "bg-mention-bg shadow-[inset_3px_0_0_var(--color-mention-bar)]"
            : isBookmarked
              ? "bg-bookmark-bg shadow-[inset_3px_0_0_var(--color-bookmark-bar)]"
              : message.pin && "bg-pin-bg shadow-[inset_3px_0_0_var(--color-pin-bar)]",
        "transition-colors motion-reduce:transition-none",
      )}
    >
      {isApp ? (
        <span className="mt-0.5 self-start">{avatar}</span>
      ) : (
        <Button
          aria-label={t("message.profileOf", { name: displayName })}
          onPress={openProfile}
          className={`mt-0.5 self-start rounded-md ${focusRing}`}
        >
          {avatar}
        </Button>
      )}

      <div className="flex min-w-0 flex-1 flex-col gap-[5px]">
        {message.pin && (
          <span className="-mb-0.5 inline-flex items-center gap-1 self-start text-[11px] font-semibold text-accent-text [&_svg]:size-3">
            <IconPin aria-hidden />
            {t("pin.label", {
              name:
                message.pin.pinnedBy?.id === myId
                  ? t("reaction.names.you")
                  : (message.pin.pinnedBy?.displayName ?? ""),
            })}
          </span>
        )}
        <div className="flex flex-wrap items-baseline gap-[7px] leading-[1.3]">
          {isApp ? (
            <>
              <span className="text-sm font-bold text-text">{displayName}</span>
              <Badge tone="tag" className="self-center">
                {t(message.isOfficial ? "app.official" : "app.tag")}
              </Badge>
            </>
          ) : (
            <Button
              onPress={openProfile}
              className={`cursor-pointer rounded-sm text-sm font-bold text-text data-hovered:underline data-hovered:underline-offset-2 ${focusRing}`}
            >
              {displayName}
            </Button>
          )}
          <MessageTime date={createdAt} />
          {channelChip}
          {message.editedAt && !message.isDeleted && (
            <span className="text-[11px] text-subtle">{t("message.edited")}</span>
          )}
          {isBookmarked && (
            <IconBookmarkFilled
              aria-label={t("message.actions.bookmark")}
              className="size-3 self-center text-accent-text"
            />
          )}
        </div>

        {message.isDeleted ? (
          <p className="m-0 text-body text-muted italic">
            {message.deletedBy
              ? t("message.deletedBy", { name: message.deletedBy.displayName })
              : t("message.deleted")}
          </p>
        ) : isEditing ? (
          <MessageEditor
            initialBody={message.body}
            onSave={(body) => handleEdit(message.id, body)}
            onClose={() => {
              setIsEditing(false);
            }}
          />
        ) : (
          <MessageContent message={message} />
        )}

        {!message.isDeleted && message.location && (
          <MessageLocationCard location={message.location} />
        )}
        {!message.isDeleted && message.poll && (
          <MessagePollCard poll={message.poll} isAuthor={message.userId === myId} />
        )}
        {!message.isDeleted && <MessageAttachments message={message} />}

        <ReactionList
          messageId={message.id}
          reactions={message.reactions}
          onOpenList={setReactionTab}
        />

        {threadMetadata && threadMetadata.replyCount > 0 && (
          <ThreadMetadataPreview
            metadata={threadMetadata}
            onPress={() => {
              onCreateThread(message.id);
            }}
          />
        )}
      </div>

      {showToolbar && (
        <MessageToolbar
          actions={actions}
          isBookmarked={isBookmarked}
          onToggleBookmark={toggleBookmark}
          onReplyInThread={() => {
            onCreateThread(message.id);
          }}
          onReact={react}
          onOverlayOpenChange={setIsOverlayOpen}
        />
      )}

      {isMobile && isSheetOpen && (
        <MessageActionSheet
          onClose={() => {
            void navigate({ search: closeDialog, to: "." });
          }}
          message={message}
          actions={actions}
          onReact={react}
        />
      )}

      <ReactionsDialog message={message} tab={reactionTab} onTabChange={setReactionTab} />

      <AlertDialog
        isOpen={isDeleteOpen}
        onOpenChange={setIsDeleteOpen}
        title={t("message.delete.title")}
        confirmLabel={t("message.delete.confirm")}
        tone="danger"
        isPending={isDeleting}
        onConfirm={() => {
          void handleDelete(message).then(() => {
            setIsDeleteOpen(false);
          });
        }}
      >
        {t("message.delete.body")}
      </AlertDialog>
    </div>
  );
};
