import { IconChevronDown } from "@tabler/icons-react";
import { useAtom } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles";
import { transitions } from "#/lib/motion";
import { collapsedChannelsAtom } from "#/providers/store/ui";

import { lastSegment } from "../utils/channelPath";
import { sumUnread } from "../utils/channelTree";
import { ChannelName } from "./ChannelName";
import { ChannelNavItem } from "./ChannelNavItem";

import type { ChannelTreeNode } from "../utils/channelTree";

type ChannelTreeItemProps = {
  workspaceId: string;
  node: ChannelTreeNode;
  depth: number;
  isLast: boolean;
};

const INDENT = 18;

// ツリーの 1 行と子孫。折りたたむと子孫の未読をこの行にまとめて出す
export const ChannelTreeItem = ({ workspaceId, node, depth, isLast }: ChannelTreeItemProps) => {
  const { t } = useTranslation();
  const [collapsed, setCollapsed] = useAtom(collapsedChannelsAtom);
  const { channel, children } = node;
  const hasChildren = children.length > 0;
  const isCollapsed = hasChildren && (collapsed[channel.id] ?? false);
  const unread = isCollapsed ? sumUnread(node) : channel;
  const lineLeft = depth * INDENT - 2;

  return (
    <div className="relative">
      {depth > 0 && (
        <>
          <span
            aria-hidden
            className={cn(
              "absolute top-0 w-px bg-(--tree-line)",
              isLast ? "h-[calc(var(--nav-row,30px)/2)]" : "-bottom-px",
            )}
            style={{ left: lineLeft }}
          />
          <span
            aria-hidden
            className="absolute top-[calc(var(--nav-row,30px)/2)] h-px w-2 bg-(--tree-line)"
            style={{ left: lineLeft }}
          />
        </>
      )}
      <div
        className={cn(
          "relative",
          hasChildren && "[&_a]:pr-7",
          !channel.isMember && "[&_a]:text-(--nav-muted)",
        )}
        style={{ paddingLeft: depth * INDENT }}
      >
        <ChannelNavItem
          workspaceId={workspaceId}
          channelId={channel.id}
          isStarred={channel.isStarred}
          isMuted={channel.isMuted}
          unreadCount={unread.unreadCount}
          showsBadge={unread.hasMention}
        >
          <ChannelName name={lastSegment(channel.name)} isPrivate={channel.isPrivate} />
          {isCollapsed && (
            <span
              aria-label={t("channel.tree.childCount", { count: children.length })}
              className="font-mono text-[10.5px] font-medium text-(--nav-muted)"
            >
              {children.length}
            </span>
          )}
        </ChannelNavItem>
        {hasChildren && (
          <Button
            aria-label={t(isCollapsed ? "channel.tree.expand" : "channel.tree.collapse", {
              name: lastSegment(channel.name),
            })}
            aria-expanded={!isCollapsed}
            onPress={() => {
              setCollapsed({ ...collapsed, [channel.id]: !isCollapsed });
            }}
            className={`absolute top-1/2 right-1.5 grid size-4 -translate-y-1/2 cursor-pointer place-items-center rounded-sm text-(--nav-muted) data-hovered:bg-(--nav-hover) max-md:size-6 ${focusRing}`}
          >
            <IconChevronDown
              aria-hidden
              className={cn(
                "size-[13px] transition-transform motion-reduce:transition-none",
                isCollapsed && "-rotate-90",
              )}
            />
          </Button>
        )}
      </div>
      <AnimatePresence initial={false}>
        {hasChildren && !isCollapsed && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={transitions.base}
            className="flex flex-col gap-px overflow-hidden"
          >
            {children.map((child, index) => (
              <ChannelTreeItem
                key={child.channel.id}
                workspaceId={workspaceId}
                node={child}
                depth={depth + 1}
                isLast={index === children.length - 1}
              />
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};
