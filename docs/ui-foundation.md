# UI 基盤

Tailwind CSS v4 + React Aria Components による UI 基盤の設計です。デザインゴールは North Star モック（issue #11）です。

## 構成

| 場所 | 役割 |
| --- | --- |
| `packages/design-tokens` (`@chat/design-tokens`) | テーマ生成（`buildTokens`）、プリセット、余白・角丸・文字・影・モーションのトークン。DOM 非依存 |
| `packages/i18n` (`@chat/i18n`) | ja / en 辞書、`createI18n`、日付・時刻・相対時刻・曜日のフォーマッタ。DOM 非依存 |
| `frontend/src/lib/theme.ts` | トークンを CSS 変数（`--c-*` など）に展開する |
| `frontend/src/styles/globals.css` | `@theme inline` で CSS 変数を Tailwind のユーティリティに割り当てる |
| `frontend/src/providers/theme/ThemeProvider.tsx` | 設定を DOM・i18n・React Aria・Motion に反映する |
| `frontend/src/components/ui/` | React Aria Components ベースの基本部品 |
| `frontend/src/pages/UiGalleryPage.tsx` | 開発時だけ開ける見本（`/dev/ui`）。テーマ・モード・言語の切り替えもここでできる |

packages は React Native と共有する前提のため、`lib` に DOM を含めない tsconfig で型検査している。ビルドはせず、TS のソースを `exports` で直接公開して Vite に処理させる。

## トークンとテーマ

- テーマ入力は `{ hue, chroma, sidebar }` の 3 値だけ。プリセットは Jade / Cobalt / Graphite / Plum。
- `buildTokens(theme, mode)` が OKLCH で計算し、sRGB に収まるまで彩度を落として hex を返す。本文・補足・サイドバーのコントラスト比はテストで保証している。
- 部品は semantic token だけを参照する。Tailwind では `bg-surface` `text-muted` `border-border` `bg-accent` `text-accent-fg` `bg-side-active` のように使う。
  - 文字サイズは `text-title` `text-body` `text-body-strong` `text-label` `text-caption` `text-mono`（サイズ・行高・太さを含む）。
  - 角丸は `rounded-sm/md/lg/xl/full`（4 / 8 / 12 / 14 / 999px）、影は `shadow-sm/md/lg/xl`。
  - 余白は Tailwind の既定（4px 刻み）がトークンと一致するのでそのまま使う。
- トークンの正は JS オブジェクト。`ThemeProvider` が実行時に `:root` へ `--c-*` / `--r-*` / `--t-*` / `--sh-*` / `--ff-*` を書き込み、`globals.css` の `@theme inline` がそれを参照する。トークンを増やしたら `globals.css` にも追加する（`lib/theme.spec.ts` が書き漏れを検出する）。
- ダークモードの Tailwind バリアントは `dark:`（`[data-mode="dark"]` を見る）。基本はトークンが切り替わるので使う必要はない。
- Tailwind の既定パレット（`gray-50` など）は `@theme { --color-*: initial; }` で無効化している。例外はスイッチのつまみに使う `white` だけ。

## 設定の保存

- 設定は `Preferences = { theme, mode, locale }`。`providers/store/preferences.ts` の `preferencesAtom` が端末（localStorage）にも保持し、ログイン前や起動直後も前回の設定で描画する。
- アカウントには `UserService.UpdatePreferences` で保存し、`GetMe` の `user.preferences` で読む。DB は `user` テーブルの列（`theme_hue` / `theme_chroma` / `theme_sidebar` / `color_mode` / `locale`）。
- `useSyncPreferences()`（`/app` のレイアウトで呼ぶ）がアカウントの設定で端末の設定を上書きする。変更は `useUpdatePreferences()(patch)` を使う。即座に反映し、保存に失敗したら元に戻してトーストを出す。
- 色相は整数で保存する（0〜359）。設定 UI は #12 で作る。

## ベーススタイルとフォント

- Tailwind の preflight を読み込み、`body` に `bg` / `text` トークンの色を付けている。
- IBM Plex Sans JP / IBM Plex Mono は `@fontsource` でセルフホストする（外部 CDN に依存せず、オフラインの PWA でも表示できるようにするため）。`unicode-range` で分割されており、使う文字のファイルだけが読み込まれる。読み込んだ woff2 は Service Worker が CacheFirst で保持する。

## ui コンポーネントの方針

- 見た目は Tailwind のクラスで付け、状態は React Aria の data 属性で書き分ける（`data-hovered:` `data-pressed:` `data-focus-visible:` `data-selected:` `data-disabled:` `data-invalid:`）。`hover:` 疑似クラスは使わない。
- 利用側の `className` は `cn`（トークン名を理解する tailwind-merge）で既定のクラスとマージされ、後勝ちで上書きできる。
- フォーカスリングは `focusRing`、入力欄・ポップオーバー・リスト項目は `styles.ts` の共通クラスを使う。
- 選択肢を扱う部品（Select / ComboBox）は `{ value, label }[]` を受け取り、`onChange` に型付きの値を返す。React Aria の `Key` を直接扱わない。
- 部品内の文言は i18n 辞書（`ui.*` / `common.*`）から取る。

## i18n

- i18next + react-i18next。辞書は `packages/i18n/src/locales/{ja,en}/<名前空間>.ts` に機能ごとに分け、`index.ts` で束ねる（並行開発で同じファイルを編集しないため）。日本語辞書が正で、英語辞書は `Messages` 型で同じキーを持つことを強制している。`t()` のキーは型検査される（`src/i18next.d.ts`）。
- キーは `機能.文脈.項目` の camelCase（例: `message.actions.delete`、`channel.create.title`）。部品共通は `ui.*`、汎用の動詞は `common.*`。
- 変数は `{{name}}`。複数形は使わず、英語でも数に依存しない言い回しにする（辞書のキーを揃えるため）。
- 日時は `formatDate` / `formatTime` / `formatDateTime` / `formatWeekday` / `formatRelativeTime(date, now, locale)`（`@chat/i18n`）を使い、`toLocaleString` を直接呼ばない。`locale` は `preferencesAtom` から取る。
- 言語を切り替えると `ThemeProvider` が `i18n.changeLanguage`、`<html lang>`、React Aria の `I18nProvider` を更新する。

## モーション

- Motion（`motion/react`）を使う。`ThemeProvider` が `MotionConfig reducedMotion="user"` を設定しているので、OS の「視差効果を減らす」で移動を伴うアニメーションは止まる。
- 時間とスプリングは `#/lib/motion` の `transitions.fast / base / spring / sheet / push` から選び、部品に直接書かない。
  - fast: フェード・ツールチップ / base: ツリーの開閉・タブ内容 / spring: ポップオーバー・選択ピル・リアクション / sheet: シート・モバイルのダイアログ / push: モバイルの画面遷移
- トークンの値は ms とスプリング係数で持ち、React Native では同じ値を Reanimated（`withTiming` / `withSpring`）に渡す。
- 単純な出入り（ポップオーバー・ツールチップ）は React Aria の `data-entering` / `data-exiting` と CSS アニメーション（`animate-pop-in` など）で行う。CSS の `transition` を書くときは `motion-reduce:` で止める。
- タブの下線は `layoutId` で移動する。`Tabs` が `LayoutGroup` でタブ群ごとに ID を分けている。

## コードハイライト

shiki を JavaScript 正規表現エンジンで使う（WASM を読み込まない）。言語はチャットでよく使うものに絞り、必要になった時点で読み込む（`features/message/utils/highlight.ts`）。色は `--shiki-light` / `--shiki-dark` を `globals.css` で表示モードに応じて切り替える。

## 部品の対応表

| 用途 | 部品 |
| --- | --- |
| ボタン / アイコンボタン | `Button`（`variant`: primary / secondary / ghost / danger、`size`: md / sm、`isPending`）/ `IconButton`（`label` 必須。aria-label とツールチップを兼ねる） |
| リンク | `LinkButton` / `Link`（`to` / `params` は TanStack Router と同じ） |
| テキスト入力 | `TextField`（`type="password"` など）/ `TextArea`。`label` 必須、`description`、`errorMessage` |
| 選択 / 補完 | `Select` / `ComboBox`（`options: { value, label }[]`、`value`、`onChange(value)`） |
| スイッチ / チェックボックス | `Switch` / `Checkbox`（`isSelected`、`onChange(boolean)`、子にラベル） |
| タブ | `Tabs` + `TabList` + `Tab` + `TabPanel`（`id` で対応付け） |
| メニュー | `Menu`（`trigger` に IconButton など）+ `MenuItem`（`icon`、`shortcut`、`tone="danger"`、`onAction`）/ `MenuSeparator` / `MenuSection`。右クリックは同じ項目を `ContextMenu` の `menu` に渡す |
| ポップオーバー | `DialogTrigger`（react-aria-components）の中に `Button` と `Popover` |
| ツールチップ | `Tooltip`（`content`、子は React Aria の Button などフォーカスできる要素） |
| ダイアログ | `Dialog`（`isOpen` / `onOpenChange` / `title` / `footer` / `size`）。モバイルでは自動で全画面シート |
| 確認ダイアログ | `AlertDialog`（`confirmLabel` / `onConfirm` / `tone="danger"`） |
| 通知 | `toast(title, { description, tone })` |
| アバター | `Avatar`（`name` / `src` / `size` / `presence`）/ グループ DM は `GroupAvatar`（`count`） |
| バッジ | `Badge`（`tone`: count / tag / accent） |
| スライダー / セグメント | `Slider`（`onChange` は動かしている間、`onChangeEnd` は確定時）/ `SegmentedControl`（`options`、`value`、`onChange`） |
| リンクのメニュー項目 | `MenuItemLink`（`to` / `params`、`target="_blank"` で新しいタブ） |
| 空の画面の案内 | `EmptyState`（`icon` / `title` / `description`） |
| 読み込み中 | `Skeleton`（`className` で大きさと形を指定）/ `Button isPending` |
| レイアウト・文字・面 | Tailwind（`flex flex-col gap-2`、`text-caption text-muted`、`rounded-lg border border-border bg-surface`） |
| フォーム | React Aria のフォーム（`<Form>`、`validationErrors`）+ zod |
| メディアクエリ | `#/lib/useMediaQuery`（`useIsMobile`） |
| コードハイライト | shiki（上記） |
