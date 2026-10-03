import { useState } from "react";

import { formatRelativeTime } from "@chat/i18n/format";
import { useMutation } from "@connectrpc/connect-query";
import { IconChartBar, IconCheck } from "@tabler/icons-react";
import { Button as AriaButton } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Badge } from "#/components/ui/Badge/Badge";
import { Button } from "#/components/ui/Button/Button";
import { cn, focusRing } from "#/components/ui/styles/styles";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { useInvalidateMessageLists } from "#/features/message/hooks/useInvalidateMessageLists";
import { PollMode } from "#/gen/chat/v1/message_pb";
import { PollService } from "#/gen/chat/v1/poll_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";
import { toastError } from "#/lib/toastError";

import type { Poll, PollOption } from "#/gen/chat/v1/message_pb";

type MessagePollCardProps = {
  poll: Poll;
  // 締め切りボタンを出すかどうか
  isAuthor: boolean;
};

const MAX_VOTER_AVATARS = 5;

/** メッセージに付けた投票。押した選択肢に投票し、もう一度押すと取り消す */
export const MessagePollCard = ({ poll, isAuthor }: MessagePollCardProps) => {
  const { t } = useTranslation();
  const { formatDateTime, formatDateWithWeekday, formatTime, locale } = useDateFormat();
  const { member } = useMentionDirectory();
  const nameOf = useDisplayName();
  // 集計の変化はメッセージの更新として WebSocket で届く
  const invalidateMessageLists = useInvalidateMessageLists();
  const close = useMutation(PollService.method.closePoll, {
    onError: toastError,
    onSuccess: invalidateMessageLists,
  });
  const vote = useMutation(PollService.method.vote, { onSuccess: invalidateMessageLists });
  // 配信されるメッセージには自分の投票が含まれないため、読み込んだ時点と投票した結果を覚えておく
  const [myOptionIds, setMyOptionIds] = useState(poll.myOptionIds);
  const maxVotes = Math.max(1, ...poll.options.map((option) => option.voteCount));

  const labelOf = (option: PollOption) => {
    if (poll.mode !== PollMode.DATE || option.startsAt === undefined) {
      return option.label;
    }
    const at = toDate(option.startsAt);
    return option.allDay
      ? formatDateWithWeekday(at)
      : `${formatDateWithWeekday(at)} ${formatTime(at)}`;
  };

  const choose = (optionId: string) => {
    const isSelected = myOptionIds.includes(optionId);
    const next = poll.allowMultiple
      ? isSelected
        ? myOptionIds.filter((id) => id !== optionId)
        : [...myOptionIds, optionId]
      : isSelected
        ? []
        : [optionId];
    vote.mutate(
      { optionIds: next, pollId: poll.id },
      {
        onError: () => {
          toast(t("poll.failed"), { tone: "danger" });
        },
        onSuccess: (res) => {
          setMyOptionIds(res.message?.poll?.myOptionIds ?? next);
        },
      },
    );
  };

  return (
    <section
      aria-label={t("poll.label")}
      className="flex w-110 max-w-full flex-col gap-2 rounded-lg border border-border bg-surface px-3 py-2.5 font-sans"
    >
      <header className="flex flex-wrap items-center gap-1.5">
        <IconChartBar aria-hidden className="size-4 shrink-0 text-accent-text" />
        <b className="min-w-0 flex-1 text-body font-bold break-words">{poll.question}</b>
        {poll.allowMultiple && <Badge tone="tag">{t("poll.multipleBadge")}</Badge>}
        {poll.anonymous && <Badge tone="tag">{t("poll.anonymousBadge")}</Badge>}
        {poll.isClosed && <Badge tone="tag">{t("poll.closedBadge")}</Badge>}
      </header>
      <ul className="m-0 flex list-none flex-col gap-1.5 p-0">
        {poll.options.map((option) => {
          const isMine = myOptionIds.includes(option.id);
          return (
            <li key={option.id}>
              <AriaButton
                isDisabled={poll.isClosed || vote.isPending}
                aria-pressed={isMine}
                onPress={() => {
                  choose(option.id);
                }}
                className={cn(
                  "relative flex w-full cursor-pointer items-center gap-2 overflow-hidden rounded-md border px-2.5 py-1.5 text-left text-body-sm text-text data-disabled:cursor-default",
                  isMine ? "border-accent" : "border-border data-hovered:bg-hover",
                  focusRing,
                )}
              >
                <span
                  aria-hidden
                  className="absolute inset-y-0 left-0 bg-accent-soft"
                  style={{ width: `${(option.voteCount / maxVotes) * 100}%` }}
                />
                <span className="relative grid size-4 shrink-0 place-items-center text-accent-text [&_svg]:size-3.5">
                  {isMine && <IconCheck aria-hidden />}
                </span>
                <span className="relative min-w-0 flex-1 break-words">{labelOf(option)}</span>
                {!poll.anonymous && option.voterIds.length > 0 && (
                  <span className="relative flex -space-x-1">
                    {option.voterIds.slice(0, MAX_VOTER_AVATARS).map((userId) => {
                      const voter = member(userId);
                      return (
                        <Avatar
                          key={userId}
                          name={nameOf(userId, "")}
                          src={voter?.avatarUrl}
                          size={18}
                        />
                      );
                    })}
                  </span>
                )}
                <span className="relative shrink-0 text-caption text-muted tabular-nums">
                  {t("poll.votes", { count: option.voteCount })}
                </span>
              </AriaButton>
            </li>
          );
        })}
      </ul>
      <footer className="flex flex-wrap items-center gap-x-2 text-caption text-muted">
        <span>{t("poll.voters", { count: poll.voterCount })}</span>
        {poll.closesAt && !poll.isClosed && (
          <span>
            {t("poll.closesAt", {
              relative: formatRelativeTime(toDate(poll.closesAt), new Date(), locale),
              time: formatDateTime(toDate(poll.closesAt)),
            })}
          </span>
        )}
        {isAuthor && !poll.isClosed && (
          <Button
            size="sm"
            variant="ghost"
            className="ml-auto"
            isPending={close.isPending}
            onPress={() => {
              close.mutate(
                { pollId: poll.id },
                {
                  onSuccess: () => {
                    toast(t("poll.closed"));
                  },
                },
              );
            }}
          >
            {t("poll.close")}
          </Button>
        )}
      </footer>
    </section>
  );
};
