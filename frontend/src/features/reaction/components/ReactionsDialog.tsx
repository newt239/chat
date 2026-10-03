import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { Tab } from "#/components/ui/Tab/Tab";
import { TabList } from "#/components/ui/TabList/TabList";
import { TabPanel } from "#/components/ui/TabPanel/TabPanel";
import { Tabs } from "#/components/ui/Tabs/Tabs";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { useDateFormat } from "#/hooks/useDateFormat";
import { toDate } from "#/lib/timestamp";
import { myUserIdAtom } from "#/providers/store/auth";

import { useToggleReaction } from "../hooks/useReactions";
import { ALL_REACTIONS_TAB, groupReactions } from "../utils/groupReactions";
import { ReactionEmoji } from "./ReactionEmoji";

import type { Message } from "#/gen/chat/v1/message_pb";

type ReactionsDialogProps = {
  message: Message;
  // 開いているタブ（絵文字か ALL_REACTIONS_TAB）。null のときは閉じている
  tab: string | null;
  onTabChange: (tab: string | null) => void;
};

// 誰がいつどのリアクションを付けたかの一覧。新しい順に並べる
export const ReactionsDialog = ({ message, tab, onTabChange }: ReactionsDialogProps) => {
  const { t } = useTranslation();
  const { toExcerpt } = useMentionDirectory();
  const { formatDateTime } = useDateFormat();
  const displayName = useDisplayName();
  const currentUserId = useAtomValue(myUserIdAtom);
  const toggleReaction = useToggleReaction(message);
  const groups = groupReactions(message.reactions, currentUserId);
  const rows = message.reactions.toSorted(
    (a, b) => toDate(b.createdAt).getTime() - toDate(a.createdAt).getTime(),
  );
  const selectedTab =
    tab !== null && groups.some(({ emoji }) => emoji === tab) ? tab : ALL_REACTIONS_TAB;
  const tabs = [
    { count: rows.length, id: ALL_REACTIONS_TAB },
    ...groups.map(({ emoji, count }) => ({ count, id: emoji })),
  ];

  return (
    <Dialog
      isOpen={tab !== null}
      onOpenChange={() => {
        onTabChange(null);
      }}
      title={t("reaction.list.title")}
    >
      <p className="-mt-1 mb-0 truncate text-label font-normal text-muted">
        {displayName(message.userId, message.user?.displayName ?? "")}: {toExcerpt(message.body)}
      </p>
      <Tabs
        selectedKey={selectedTab}
        onSelectionChange={(key) => {
          onTabChange(String(key));
        }}
        className="-mx-4.5 min-h-60"
      >
        <TabList aria-label={t("reaction.list.title")}>
          {tabs.map(({ id, count }) => (
            <Tab key={id} id={id}>
              <span
                className={id === ALL_REACTIONS_TAB ? "text-body-sm" : "text-base leading-none"}
              >
                {id === ALL_REACTIONS_TAB ? t("reaction.list.all") : <ReactionEmoji emoji={id} />}
              </span>
              <small className="font-mono text-caption text-subtle tabular-nums">{count}</small>
            </Tab>
          ))}
        </TabList>
        {tabs.map(({ id }) => (
          <TabPanel key={id} id={id} className="overflow-y-auto px-2.5 py-1.5">
            <ul className="m-0 flex list-none flex-col p-0">
              {rows
                .filter((row) => id === ALL_REACTIONS_TAB || row.emoji === id)
                .map((row) => {
                  const isMine = row.user?.id === currentUserId;
                  const name = displayName(row.user?.id ?? "", row.user?.displayName ?? "");
                  return (
                    <li
                      key={`${row.emoji}-${row.user?.id}`}
                      className="flex items-center gap-2.5 rounded-lg px-2 py-1.5"
                    >
                      <Avatar name={name} src={row.user?.avatarUrl} size={30} />
                      <span className="flex min-w-0 flex-1 flex-col leading-snug">
                        <span className="truncate text-body-sm font-medium">
                          {isMine ? t("reaction.names.you") : name}
                        </span>
                        <small className="font-mono text-caption text-subtle tabular-nums">
                          {formatDateTime(toDate(row.createdAt))}
                        </small>
                      </span>
                      {id === ALL_REACTIONS_TAB && (
                        <span aria-hidden className="text-lg leading-none">
                          <ReactionEmoji emoji={row.emoji} />
                        </span>
                      )}
                      {isMine && (
                        <Button
                          variant="secondary"
                          size="sm"
                          onPress={() => {
                            toggleReaction(row.emoji);
                          }}
                        >
                          {t("reaction.list.undo")}
                        </Button>
                      )}
                    </li>
                  );
                })}
            </ul>
          </TabPanel>
        ))}
      </Tabs>
    </Dialog>
  );
};
