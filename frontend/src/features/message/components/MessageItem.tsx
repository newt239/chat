import { useState } from "react";

import { formatDateTime, formatTime } from "@chat/i18n";
import { IconBookmarkFilled, IconPin } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Avatar } from "#/components/ui/Avatar";
import { cn, focusRing } from "#/components/ui/styles";
import { MessageAttachments } from "#/features/attachment/components/MessageAttachments";
import { ReactionList } from "#/features/reaction/components/ReactionList";
import { ReactionsDialog } from "#/features/reaction/components/ReactionsDialog";
import { useToggleReaction } from "#/features/reaction/hooks/useReactions";
import { ALL_REACTIONS_TAB } from "#/features/reaction/utils/reactionTabs";
import { toDate } from "#/lib/timestamp";
import { useIsMobile } from "#/lib/useMediaQuery";
import { userAtom } from "#/providers/store/auth";
import { preferencesAtom } from "#/providers/store/preferences";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import { useLongPress } from "../hooks/useLongPress";
import { useMessageActions } from "../hooks/useMessageActions";
import { useMessageMenuActions } from "../hooks/useMessageMenuActions";
import { MessageActionSheet } from "./MessageActionSheet";
import { MessageContent } from "./MessageContent";
import { MessageEditor } from "./MessageEditor";
import { MessageToolbar } from "./MessageToolbar";
import { ThreadMetadataPreview } from "./ThreadMetadataPreview";

import type { Message, ThreadMetadata } from "#/gen/chat/v1/message_pb";

type MessageItemProps = {
  message: Message;
  currentUserId: string | null;
  onCopyLink: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  threadMetadata?: ThreadMetadata;
  onOpenThread?: (messageId: string) => void;
  isHighlighted?: boolean;
};

export const MessageItem = ({
  message,
  currentUserId,
  onCopyLink,
  onCreateThread,
  threadMetadata,
  onOpenThread,
  isHighlighted = false,
}: MessageItemProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const currentUser = useAtomValue(userAtom);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);
  const isMobile = useIsMobile();
  const [isEditing, setIsEditing] = useState(false);
  const [isSheetOpen, setIsSheetOpen] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isOverlayOpen, setIsOverlayOpen] = useState(false);
  const [isFocusWithin, setIsFocusWithin] = useState(false);
  const [isHovered, setIsHovered] = useState(false);
  const [reactionTab, setReactionTab] = useState<string | null>(null);
  const { isPressed, longPressProps } = useLongPress(() => {
    setIsSheetOpen(true);
  }, isMobile && !isEditing);
  const { handleEdit, handleDelete, isDeleting } = useMessageActions();
  const toggleReaction = useToggleReaction(message.id);

  const { actions, isBookmarked, toggleBookmark } = useMessageMenuActions({
    isAuthor: message.userId === currentUserId,
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
  });

  const react = (emoji: string) => {
    toggleReaction(
      emoji,
      message.reactions.some(
        (reaction) => reaction.emoji === emoji && reaction.user?.id === currentUser?.id,
      ),
    );
  };

  const openProfile = () => {
    setRightSidePanelView({ type: "user-profile", userId: message.userId });
  };

  const displayName = message.user?.displayName ?? "";
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
        message.pin && "bg-pin-bg shadow-[inset_3px_0_0_var(--color-pin-bar)]",
        isHighlighted && "bg-accent-soft",
        "transition-colors motion-reduce:transition-none",
      )}
    >
      <Button
        aria-label={t("message.profileOf", { name: displayName })}
        onPress={openProfile}
        className={`mt-0.5 self-start rounded-md ${focusRing}`}
      >
        <Avatar name={displayName} src={message.user?.avatarUrl} size={isMobile ? 34 : 32} />
      </Button>

      <div className="flex min-w-0 flex-1 flex-col gap-[5px]">
        {message.pin && (
          <span className="-mb-0.5 inline-flex items-center gap-1 self-start text-[11px] font-semibold text-accent-text [&_svg]:size-3">
            <IconPin aria-hidden />
            {t("pin.label", {
              name:
                message.pin.pinnedBy?.id === currentUserId
                  ? t("reaction.names.you")
                  : (message.pin.pinnedBy?.displayName ?? ""),
            })}
          </span>
        )}
        <div className="flex flex-wrap items-baseline gap-[7px] leading-[1.3]">
          <Button
            onPress={openProfile}
            className={`cursor-pointer rounded-sm text-text data-hovered:underline data-hovered:underline-offset-2 ${focusRing}`}
          >
            {/* Mantine のリセットが button の font を上書きするため、中の要素で指定する */}
            <span className="text-sm font-bold">{displayName}</span>
          </Button>
          <time
            dateTime={createdAt.toISOString()}
            title={formatDateTime(createdAt, locale)}
            className="font-mono text-[11.5px] text-subtle tabular-nums"
          >
            {formatTime(createdAt, locale)}
          </time>
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

        {!message.isDeleted && <MessageAttachments message={message} />}

        <ReactionList
          messageId={message.id}
          reactions={message.reactions}
          onOpenList={setReactionTab}
        />

        {threadMetadata && threadMetadata.replyCount > 0 && onOpenThread && (
          <ThreadMetadataPreview
            metadata={threadMetadata}
            onPress={() => {
              onOpenThread(message.id);
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

      {isMobile && (
        <MessageActionSheet
          isOpen={isSheetOpen}
          onOpenChange={setIsSheetOpen}
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
          void handleDelete(message.id).then(() => {
            setIsDeleteOpen(false);
          });
        }}
      >
        {t("message.delete.body")}
      </AlertDialog>
    </div>
  );
};
