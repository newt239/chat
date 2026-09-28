import { useState } from "react";

import {
  colorTokenNames,
  findThemePreset,
  themePresetNames,
  themePresets,
} from "@chat/design-tokens";
import { formatDate, formatRelativeTime, formatTime, formatWeekday } from "@chat/i18n";
import { IconCopy, IconDots, IconPin, IconSettings, IconTrash } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { DialogTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog";
import { Avatar } from "#/components/ui/Avatar";
import { Badge } from "#/components/ui/Badge";
import { Button } from "#/components/ui/Button";
import { Checkbox } from "#/components/ui/Checkbox";
import { ComboBox } from "#/components/ui/ComboBox";
import { ContextMenu } from "#/components/ui/ContextMenu";
import { Dialog } from "#/components/ui/Dialog";
import { GroupAvatar } from "#/components/ui/GroupAvatar";
import { IconButton } from "#/components/ui/IconButton";
import { Link } from "#/components/ui/Link";
import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator";
import { Popover } from "#/components/ui/Popover";
import { Select } from "#/components/ui/Select";
import { Skeleton } from "#/components/ui/Skeleton";
import { Switch } from "#/components/ui/Switch";
import { Tab } from "#/components/ui/Tab";
import { TabList } from "#/components/ui/TabList";
import { TabPanel } from "#/components/ui/TabPanel";
import { Tabs } from "#/components/ui/Tabs";
import { TextArea } from "#/components/ui/TextArea";
import { TextField } from "#/components/ui/TextField";
import { toast } from "#/components/ui/toast";
import { Tooltip } from "#/components/ui/Tooltip";
import { CodeBlock } from "#/features/message/components/markdown/CodeBlock";
import { useUpdatePreferences } from "#/features/settings/hooks/usePreferences";
import { preferencesAtom } from "#/providers/store/preferences";

const channels = [
  { label: "general", value: "general" },
  { label: "frontend", value: "frontend" },
  { label: "frontend/web", value: "frontend-web" },
  { label: "backend", value: "backend" },
] as const;

const sampleDate = new Date(2026, 8, 28, 10, 16);

// 開発時だけ表示する ui コンポーネントとテーマの見本。画面移行の見た目確認に使う
export const UiGalleryPage = () => {
  const { t, i18n } = useTranslation();
  const preferences = useAtomValue(preferencesAtom);
  const updatePreferences = useUpdatePreferences();
  const [channel, setChannel] = useState<(typeof channels)[number]["value"] | null>("general");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isAlertOpen, setIsAlertOpen] = useState(false);
  const { locale } = preferences;
  const preset = findThemePreset(preferences.theme) ?? "custom";

  const menuItems = (
    <>
      <MenuItem
        icon={<IconPin />}
        shortcut="P"
        onAction={() => {
          toast("ピン留めしました");
        }}
      >
        チャンネルにピン留め
      </MenuItem>
      <MenuItem
        icon={<IconCopy />}
        onAction={() => {
          toast("リンクをコピーしました");
        }}
      >
        リンクをコピー
      </MenuItem>
      <MenuSeparator />
      <MenuItem
        icon={<IconTrash />}
        tone="danger"
        onAction={() => {
          setIsAlertOpen(true);
        }}
      >
        メッセージを削除
      </MenuItem>
    </>
  );

  return (
    <div className="min-h-full bg-bg font-sans text-body text-text">
      <div className="mx-auto flex max-w-5xl flex-col gap-6 px-4 py-6">
        <header className="flex flex-col gap-1">
          <h1 className="m-0 text-xl font-bold">UI gallery</h1>
          <p className="m-0 text-muted">
            {formatDate(sampleDate, locale)} ({formatWeekday(sampleDate, locale)}){" "}
            {formatTime(sampleDate, locale)} ·{" "}
            {formatRelativeTime(new Date(sampleDate.getTime() - 180_000), sampleDate, locale)} ·{" "}
            <Link to="/app">/app</Link>
          </p>
        </header>

        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 md:grid-cols-4">
          <Select
            label={t("preferences.theme.title")}
            options={[
              ...themePresetNames.map((name) => ({
                label: t(`preferences.theme.presets.${name}`),
                value: name,
              })),
              { label: t("preferences.theme.custom"), value: "custom" as const },
            ]}
            value={preset}
            onChange={(name) => {
              if (name !== "custom") {
                updatePreferences({ theme: themePresets[name] });
              }
            }}
          />
          <Select
            label={t("preferences.mode.title")}
            options={(["light", "dark", "system"] as const).map((mode) => ({
              label: t(`preferences.mode.${mode}`),
              value: mode,
            }))}
            value={preferences.mode}
            onChange={(mode) => {
              updatePreferences({ mode });
            }}
          />
          <Select
            label={t("preferences.locale.title")}
            options={(["ja", "en"] as const).map((value) => ({
              label: t(`preferences.locale.${value}`),
              value,
            }))}
            value={locale}
            onChange={(value) => {
              updatePreferences({ locale: value });
            }}
          />
          <Switch
            className="self-end pb-2"
            isSelected={preferences.theme.sidebar === "tinted"}
            onChange={(tinted) => {
              updatePreferences({
                theme: { ...preferences.theme, sidebar: tinted ? "tinted" : "light" },
              });
            }}
          >
            {t("preferences.theme.sidebar.title")}: {t("preferences.theme.sidebar.tinted")}
          </Switch>
          <div className="flex flex-wrap items-center gap-1 md:col-span-4">
            <span className="mr-2 text-label">{t("preferences.theme.hue")}</span>
            {Array.from({ length: 12 }, (_, i) => i * 30).map((hue) => (
              <Button
                key={hue}
                size="sm"
                variant={preferences.theme.hue === hue ? "primary" : "secondary"}
                onPress={() => {
                  updatePreferences({ theme: { ...preferences.theme, hue } });
                }}
              >
                {hue}
              </Button>
            ))}
          </div>
        </section>

        <section className="flex flex-wrap gap-2 rounded-lg border border-border bg-surface p-4">
          {colorTokenNames.map((name) => (
            <div key={name} className="flex w-40 items-center gap-2 text-caption">
              <span
                className="size-6 shrink-0 rounded-sm border border-border"
                style={{ background: `var(--c-${name})` }}
              />
              {name}
            </div>
          ))}
        </section>

        <section className="flex flex-wrap items-center gap-2 rounded-lg border border-border bg-surface p-4">
          <Button
            onPress={() => {
              toast("保存しました", { tone: "success" });
            }}
          >
            Primary
          </Button>
          <Button
            variant="secondary"
            onPress={() => {
              toast("通知です");
            }}
          >
            Secondary
          </Button>
          <Button variant="ghost">Ghost</Button>
          <Button
            variant="danger"
            onPress={() => {
              toast("失敗しました", { tone: "danger" });
            }}
          >
            Danger
          </Button>
          <Button isDisabled>Disabled</Button>
          <Button isPending>{t("common.save")}</Button>
          <Button size="sm" variant="secondary">
            Small
          </Button>
          <IconButton label={t("common.close")}>
            <IconSettings />
          </IconButton>
          <Menu
            trigger={
              <IconButton label="その他">
                <IconDots />
              </IconButton>
            }
          >
            {menuItems}
          </Menu>
          <DialogTrigger>
            <Button variant="secondary">Popover</Button>
            <Popover aria-label="プロフィール" className="w-72 p-4">
              <p className="m-0 text-muted">ポップオーバーの中身</p>
            </Popover>
          </DialogTrigger>
          <Tooltip content="ツールチップ">
            <Button variant="ghost">Tooltip</Button>
          </Tooltip>
          <Button
            variant="secondary"
            onPress={() => {
              setIsDialogOpen(true);
            }}
          >
            Dialog
          </Button>
          <Button
            variant="danger"
            onPress={() => {
              setIsAlertOpen(true);
            }}
          >
            AlertDialog
          </Button>
        </section>

        <section className="grid gap-4 rounded-lg border border-border bg-surface p-4 md:grid-cols-2">
          <TextField
            label="チャンネル名"
            description="小文字の英数字とハイフン"
            placeholder="frontend"
          />
          <TextField label="メールアドレス" errorMessage="メールアドレスの形式が正しくありません" />
          <TextArea label="説明" placeholder="このチャンネルの目的" />
          <div className="flex flex-col gap-3">
            <ComboBox
              label="チャンネル"
              options={channels}
              value={channel}
              onChange={setChannel}
              placeholder="検索"
            />
            <Checkbox defaultSelected>メンションを通知する</Checkbox>
            <Checkbox isIndeterminate>一部のチャンネル</Checkbox>
            <Switch defaultSelected>コンパクト表示</Switch>
          </div>
        </section>

        <section className="rounded-lg border border-border bg-surface">
          <Tabs>
            <TabList aria-label="タブ">
              <Tab id="messages">メッセージ</Tab>
              <Tab id="files">ファイル</Tab>
              <Tab id="pins">ピン留め</Tab>
            </TabList>
            <TabPanel id="messages" className="p-4">
              <ContextMenu aria-label="メッセージの操作" menu={menuItems}>
                <div className="rounded-md border border-dashed border-border-strong p-6 text-center text-muted">
                  右クリックでコンテキストメニュー
                </div>
              </ContextMenu>
            </TabPanel>
            <TabPanel id="files" className="p-4">
              <CodeBlock>
                <code className="language-ts">
                  {"export const space = { 1: 4, 2: 8 } as const;\nconst hue: number = 168;\n"}
                </code>
              </CodeBlock>
            </TabPanel>
            <TabPanel id="pins" className="flex flex-col gap-2 p-4">
              <Skeleton className="h-4 w-48" />
              <Skeleton className="h-4 w-72" />
              <Skeleton className="size-8 rounded-full" />
            </TabPanel>
          </Tabs>
        </section>

        <section className="flex flex-wrap items-center gap-3 rounded-lg border border-border bg-surface p-4">
          <Avatar name="田中 美咲" presence="online" />
          <Avatar name="Kenta" presence="away" size={40} />
          <Avatar name="Ren" presence="offline" size={24} />
          <GroupAvatar count={4} />
          <Badge>3</Badge>
          <Badge tone="tag">BOT</Badge>
          <Badge tone="accent">@frontend</Badge>
          <span className="rounded-md bg-side px-3 py-2 text-side-fg [--dot-ring:var(--c-side)]">
            <Avatar name="Yui" presence="online" size={18} /> sidebar
          </span>
        </section>
      </div>

      <Dialog
        isOpen={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        title="チャンネルを作成"
        footer={
          <>
            <Button
              variant="secondary"
              onPress={() => {
                setIsDialogOpen(false);
              }}
            >
              {t("common.cancel")}
            </Button>
            <Button
              onPress={() => {
                setIsDialogOpen(false);
              }}
            >
              {t("common.save")}
            </Button>
          </>
        }
      >
        <TextField label="チャンネル名" />
        <p className="m-0 text-muted">言語: {i18n.language}</p>
      </Dialog>
      <AlertDialog
        isOpen={isAlertOpen}
        onOpenChange={setIsAlertOpen}
        title="メッセージを削除しますか？"
        confirmLabel={t("common.delete")}
        tone="danger"
        onConfirm={() => {
          setIsAlertOpen(false);
          toast("メッセージを削除しました");
        }}
      >
        元に戻せません。スレッドの返信も削除されます。
      </AlertDialog>
    </div>
  );
};
