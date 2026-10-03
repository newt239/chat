import { IconHash, IconLock, IconUsers } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Link } from "#/components/ui/Link/Link";
import { MemberRow } from "#/features/member/components/MemberRow";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { workspaceRoleKeys } from "#/features/member/utils/workspaceRoleKeys";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { MessageListCard } from "#/features/message/components/MessageListCard";
import { excerpt } from "#/features/search/utils/excerpt";
import { splitHighlights } from "#/features/search/utils/splitHighlights";

import { SearchResultSection } from "./SearchResultSection";

import type { SearchFilter } from "#/features/search/schemas";
import type { Channel } from "#/gen/chat/v1/channel_service_pb";
import type { MessageSearchHit } from "#/gen/chat/v1/search_service_pb";
import type { UserGroup } from "#/gen/chat/v1/user_group_service_pb";
import type { WorkspaceMember } from "#/gen/chat/v1/workspace_service_pb";

type SearchResultListProps = {
  messages: MessageSearchHit[];
  channels: Channel[];
  users: WorkspaceMember[];
  groups: UserGroup[];
  filter: SearchFilter;
  workspaceId: string;
};

export const SearchResultList = ({
  messages,
  channels,
  users,
  groups,
  filter,
  workspaceId,
}: SearchResultListProps) => {
  const { t } = useTranslation();
  const { toText } = useMentionDirectory();
  const displayName = useDisplayName();
  const shows = (section: SearchFilter) => filter === "all" || filter === section;

  return (
    <div className="flex flex-col gap-1 pb-4 font-sans text-text">
      {shows("messages") && messages.length > 0 && (
        <SearchResultSection title={t("search.sections.messages")}>
          {messages.map(({ message, highlights }) => {
            if (message === undefined) {
              return null;
            }
            const authorName = displayName(message.userId, message.user?.displayName ?? "");
            const body = excerpt(toText(message.body), highlights, 40);
            return (
              <div key={message.id} className="mx-4.5 my-1.5">
                <MessageListCard workspaceId={workspaceId} message={message}>
                  <div className="flex gap-2.5 px-3 py-2">
                    <Avatar name={authorName} src={message.user?.avatarUrl} size={32} />
                    <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                      <span className="text-body-strong">{authorName}</span>
                      <p className="m-0 line-clamp-3 text-body break-words whitespace-pre-wrap">
                        {body.isTrimmed && "…"}
                        {splitHighlights(body.text, body.ranges).map((part) =>
                          part.isMatch ? (
                            <mark
                              key={part.start}
                              className="rounded-xs bg-mention-chip text-inherit"
                            >
                              {part.text}
                            </mark>
                          ) : (
                            part.text
                          ),
                        )}
                      </p>
                    </div>
                  </div>
                </MessageListCard>
              </div>
            );
          })}
        </SearchResultSection>
      )}

      {shows("channels") && channels.length > 0 && (
        <SearchResultSection title={t("search.sections.channels")}>
          <ul className="m-0 flex list-none flex-col px-2.5 py-0">
            {channels.map((channel) => (
              <li key={channel.id}>
                <Link
                  to="/app/$workspaceId/$channelId"
                  params={{ channelId: channel.id, workspaceId }}
                  className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-body-sm text-text no-underline data-hovered:bg-hover"
                >
                  {channel.isPrivate ? (
                    <IconLock aria-hidden className="size-4 shrink-0 text-muted" />
                  ) : (
                    <IconHash aria-hidden className="size-4 shrink-0 text-muted" />
                  )}
                  <span className="flex min-w-0 flex-1 flex-col leading-snug">
                    <span className="truncate">{channel.name}</span>
                    <small className="truncate text-xs text-muted">
                      {channel.description || t("search.noDescription")}
                    </small>
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        </SearchResultSection>
      )}

      {shows("users") && users.length > 0 && (
        <SearchResultSection title={t("search.sections.users")}>
          <ul className="m-0 flex list-none flex-col px-2.5 py-0">
            {users.map((user) => (
              <li key={user.userId}>
                <MemberRow
                  userId={user.userId}
                  name={displayName(user.userId, user.displayName)}
                  avatarUrl={user.avatarUrl}
                  detail={`${t(workspaceRoleKeys[user.role])} · ${user.email}`}
                />
              </li>
            ))}
          </ul>
        </SearchResultSection>
      )}

      {shows("groups") && groups.length > 0 && (
        <SearchResultSection title={t("search.sections.groups")}>
          <ul className="m-0 flex list-none flex-col px-2.5 py-0">
            {groups.map((group) => (
              <li
                key={group.id}
                className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-body-sm"
              >
                <IconUsers aria-hidden className="size-4 shrink-0 text-muted" />
                <span className="flex min-w-0 flex-1 flex-col leading-snug">
                  <span className="truncate font-semibold text-accent-text">@{group.name}</span>
                  <small className="truncate text-xs text-muted">
                    {group.description || t("search.noDescription")}
                  </small>
                </span>
              </li>
            ))}
          </ul>
        </SearchResultSection>
      )}
    </div>
  );
};
