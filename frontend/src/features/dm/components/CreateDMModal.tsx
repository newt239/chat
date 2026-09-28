import { useId, useState } from "react";

import { IconCheck, IconLock, IconSearch, IconX } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import {
  Button as AriaButton,
  Form,
  Input,
  ListBox,
  ListBoxItem,
  SearchField,
  Text,
} from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar";
import { Button } from "#/components/ui/Button";
import { Dialog } from "#/components/ui/Dialog";
import { cn, fieldStyles, focusRing } from "#/components/ui/styles";
import { toast } from "#/components/ui/toast";
import { ChannelNameField } from "#/features/channel/components/ChannelNameField";
import { useChannels, useCreateChannel } from "#/features/channel/hooks/useChannel";
import { channelPathErrorKeys, validateChannelPath } from "#/features/channel/utils/channelPath";
import { useMembers } from "#/features/member/hooks/useMembers";
import { transitions } from "#/lib/motion";
import { userAtom } from "#/providers/store/auth";

import { useCreateDM, useCreateGroupDM } from "../hooks/useDM";

// 自分を含めた DM の上限。超えると非公開チャンネルとして作る
const DM_MAX = 10;

type CreateDMModalProps = {
  workspaceId: string;
  opened: boolean;
  onClose: () => void;
};

export const CreateDMModal = ({ workspaceId, opened, onClose }: CreateDMModalProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const formId = useId();
  const currentUser = useAtomValue(userAtom);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [query, setQuery] = useState("");
  const [channelName, setChannelName] = useState("");
  const [isTouched, setIsTouched] = useState(false);

  const { data: members } = useMembers(opened ? workspaceId : null);
  const { data: channels } = useChannels(opened ? workspaceId : null);
  const createDM = useCreateDM();
  const createGroupDM = useCreateGroupDM();
  const createChannel = useCreateChannel();

  const candidates = (members ?? []).filter((member) => member.userId !== currentUser?.id);
  const normalizedQuery = query.trim().toLowerCase();
  const visibleCandidates = candidates.filter(
    (member) =>
      normalizedQuery.length === 0 ||
      member.displayName.toLowerCase().includes(normalizedQuery) ||
      (member.nickname?.toLowerCase().includes(normalizedQuery) ?? false) ||
      member.email.toLowerCase().includes(normalizedQuery),
  );
  const selectedMembers = selectedIds.flatMap(
    (id) => candidates.find((member) => member.userId === id) ?? [],
  );

  const total = selectedIds.length + 1;
  const isOverLimit = total > DM_MAX;
  const nameError = isOverLimit
    ? validateChannelPath(
        channelName,
        (channels ?? []).map((channel) => channel.name),
      )
    : null;
  const mutations = [createDM, createGroupDM, createChannel];
  const isPending = mutations.some((mutation) => mutation.isPending);
  const mutationError = mutations.find((mutation) => mutation.error !== null)?.error;

  const close = () => {
    setSelectedIds([]);
    setQuery("");
    setChannelName("");
    setIsTouched(false);
    for (const mutation of mutations) {
      mutation.reset();
    }
    onClose();
  };

  const openChannel = (channelId: string) => {
    close();
    void navigate({ params: { channelId, workspaceId }, to: "/app/$workspaceId/$channelId" });
  };

  const toggle = (userId: string) => {
    setSelectedIds((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId],
    );
  };

  const submit = async () => {
    setIsTouched(true);
    const [firstUserId] = selectedIds;
    if (firstUserId === undefined || nameError !== null) {
      return;
    }
    if (isOverLimit) {
      const { channel } = await createChannel.mutateAsync({
        isPrivate: true,
        memberIds: selectedIds,
        name: channelName,
        workspaceId,
      });
      toast(t("channel.create.created", { name: channelName }), { tone: "success" });
      if (channel !== undefined) {
        openChannel(channel.id);
      }
      return;
    }
    const { directMessage } =
      selectedIds.length > 1
        ? await createGroupDM.mutateAsync({ userIds: selectedIds, workspaceId })
        : await createDM.mutateAsync({ userId: firstUserId, workspaceId });
    if (directMessage !== undefined) {
      openChannel(directMessage.id);
    }
  };

  const submitLabel = isOverLimit
    ? t("dm.create.submitChannel")
    : selectedIds.length > 1
      ? t("dm.create.submitGroup")
      : t("dm.create.submitDM");

  return (
    <Dialog
      isOpen={opened}
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          close();
        }
      }}
      title={t("dm.create.title")}
      size="md"
      footer={
        <>
          <Button variant="secondary" onPress={close}>
            {t("common.cancel")}
          </Button>
          <Button
            type="submit"
            form={formId}
            isDisabled={selectedIds.length === 0}
            isPending={isPending}
          >
            {submitLabel}
          </Button>
        </>
      }
    >
      <Form
        id={formId}
        className="flex flex-col gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          void submit();
        }}
      >
        <div className="flex flex-col gap-1.5">
          <SearchField
            value={query}
            onChange={setQuery}
            aria-label={t("dm.create.search")}
            className={cn(
              fieldStyles.input,
              "flex items-center gap-2 px-2.5 data-focus-within:border-accent data-focus-within:ring-3 data-focus-within:ring-accent-soft",
            )}
          >
            <IconSearch aria-hidden className="size-4 shrink-0 text-subtle" />
            <Input
              placeholder={t("dm.create.search")}
              className="h-full min-w-0 flex-1 border-0 bg-transparent font-sans text-[13.5px] text-text outline-none placeholder:text-subtle [&::-webkit-search-cancel-button]:hidden"
            />
          </SearchField>
          {selectedMembers.length > 0 && (
            <ul className="m-0 flex list-none flex-wrap gap-1.5 p-0">
              {selectedMembers.map((member) => (
                <li
                  key={member.userId}
                  className="inline-flex items-center gap-1 rounded-full bg-accent-soft py-0.5 pr-1 pl-2.5 text-xs font-semibold text-accent-text"
                >
                  {member.nickname ?? member.displayName}
                  <AriaButton
                    aria-label={t("dm.create.removeSelected", {
                      name: member.nickname ?? member.displayName,
                    })}
                    onPress={() => {
                      toggle(member.userId);
                    }}
                    className={cn(
                      "grid size-[18px] cursor-pointer place-items-center rounded-full data-hovered:bg-accent/20",
                      focusRing,
                    )}
                  >
                    <IconX aria-hidden className="size-3" />
                  </AriaButton>
                </li>
              ))}
            </ul>
          )}
          <p className={cn("m-0 text-caption", isOverLimit ? "text-danger" : "text-muted")}>
            {t("dm.create.count", { count: total, max: DM_MAX })} ·{" "}
            {t("dm.create.limitHint", { max: DM_MAX })}
          </p>
        </div>

        <AnimatePresence initial={false}>
          {isOverLimit && (
            <motion.div
              key="over"
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: "auto", opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={transitions.base}
              className="overflow-hidden"
            >
              <div className="flex flex-col gap-2.5 rounded-lg border border-mention-bar bg-mention-bg px-3 py-2.5">
                <span className="flex items-center gap-1.5 text-body-strong">
                  <IconLock aria-hidden className="size-3.5" />
                  {t("dm.create.callout")}
                </span>
                <ChannelNameField
                  label={t("channel.create.name")}
                  prefix="#"
                  value={channelName}
                  onChange={setChannelName}
                  placeholder="release-war-room"
                  description={t("channel.create.nameHint")}
                  errorMessage={
                    isTouched && nameError !== null
                      ? t(channelPathErrorKeys[nameError], { name: channelName })
                      : null
                  }
                />
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        <ListBox
          aria-label={t("dm.create.members")}
          selectionMode="multiple"
          selectedKeys={selectedIds}
          onSelectionChange={(keys) => {
            // 選んだ順にチップを並べるため、既存の並びを保って新しく選んだ人を末尾に足す
            const next = candidates
              .filter((member) => keys === "all" || keys.has(member.userId))
              .map((member) => member.userId);
            setSelectedIds((prev) => [
              ...prev.filter((id) => next.includes(id)),
              ...next.filter((id) => !prev.includes(id)),
            ]);
          }}
          items={visibleCandidates}
          renderEmptyState={() => (
            <p className="m-0 px-2.5 py-2 text-caption text-muted">{t("dm.create.noResults")}</p>
          )}
          className="flex max-h-[210px] flex-col overflow-y-auto rounded-lg border border-border outline-none"
        >
          {(member) => (
            <ListBoxItem
              id={member.userId}
              textValue={member.nickname ?? member.displayName}
              className="flex cursor-pointer items-center gap-2.5 px-2.5 py-1.5 text-[13px] text-text outline-none data-focus-visible:bg-hover data-hovered:bg-hover"
            >
              {({ isSelected }) => (
                <>
                  <span
                    className={cn(
                      "grid size-[15px] shrink-0 place-items-center rounded-sm border border-border-strong bg-surface text-accent-fg",
                      isSelected && "border-accent bg-accent",
                    )}
                  >
                    {isSelected && <IconCheck aria-hidden stroke={3} className="size-3" />}
                  </span>
                  <Avatar name={member.displayName} src={member.avatarUrl} size={24} />
                  <Text slot="label" className="min-w-0 flex-1 truncate">
                    {member.nickname ?? member.displayName}
                  </Text>
                  <Text slot="description" className="truncate text-[11.5px] text-subtle">
                    {member.email}
                  </Text>
                </>
              )}
            </ListBoxItem>
          )}
        </ListBox>

        {isTouched && selectedIds.length === 0 && (
          <p className="m-0 text-caption text-danger">{t("dm.create.selectAtLeastOne")}</p>
        )}
        {mutationError && (
          <p role="alert" className="m-0 text-caption text-danger">
            {mutationError.message}
          </p>
        )}
      </Form>
    </Dialog>
  );
};
