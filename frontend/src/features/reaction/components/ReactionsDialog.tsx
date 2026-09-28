import { formatDateTime } from "@chat/i18n";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Button } from "#/components/ui/Button";
import { Dialog } from "#/components/ui/Dialog";
import { Tab } from "#/components/ui/Tab";
import { TabList } from "#/components/ui/TabList";
import { TabPanel } from "#/components/ui/TabPanel";
import { Tabs } from "#/components/ui/Tabs";
import { toPlainText } from "#/features/message/utils/markdown/plainText";
import { toDate } from "#/lib/timestamp";
import { userAtom } from "#/providers/store/auth";
import { preferencesAtom } from "#/providers/store/preferences";

import { useToggleReaction } from "../hooks/useReactions";
import { groupReactions } from "../utils/groupReactions";
import { ALL_REACTIONS_TAB } from "../utils/reactionTabs";

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
  const { locale } = useAtomValue(preferencesAtom);
  const currentUserId = useAtomValue(userAtom)?.id ?? null;
  const toggleReaction = useToggleReaction(message.id);
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
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onTabChange(null);
        }
      }}
      title={t("reaction.list.title")}
    >
      <p className="-mt-1 mb-0 truncate text-[12.5px] text-muted">
        {message.user?.displayName}:{" "}
        {toPlainText(message.body) || t("message.sheet.attachmentOnly")}
      </p>
      <Tabs
        selectedKey={selectedTab}
        onSelectionChange={(key) => {
          onTabChange(String(key));
        }}
        className="-mx-[18px] min-h-[240px]"
      >
        <TabList aria-label={t("reaction.list.title")}>
          {tabs.map(({ id, count }) => (
            <Tab key={id} id={id}>
              <span className={id === ALL_REACTIONS_TAB ? "text-[13px]" : "text-base leading-none"}>
                {id === ALL_REACTIONS_TAB ? t("reaction.list.all") : id}
              </span>
              <small className="font-mono text-[11px] text-subtle tabular-nums">{count}</small>
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
                  return (
                    <li
                      key={`${row.emoji}-${row.user?.id}`}
                      className="flex items-center gap-2.5 rounded-lg px-2 py-1.5"
                    >
                      <Avatar
                        name={row.user?.displayName ?? ""}
                        src={row.user?.avatarUrl}
                        size={30}
                      />
                      <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
                        <span className="truncate text-[13.5px] font-medium">
                          {isMine ? t("reaction.names.you") : row.user?.displayName}
                        </span>
                        <small className="font-mono text-[11.5px] text-subtle tabular-nums">
                          {formatDateTime(toDate(row.createdAt), locale)}
                        </small>
                      </span>
                      {id === ALL_REACTIONS_TAB && (
                        <span aria-hidden className="text-lg leading-none">
                          {row.emoji}
                        </span>
                      )}
                      {isMine && (
                        <Button
                          variant="secondary"
                          size="sm"
                          onPress={() => {
                            toggleReaction(row.emoji, true);
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
