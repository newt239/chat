# フロントエンド実装レビューレポート

**レビュー日**: 2025-10-30
**対象**: `/frontend/src`
**レビュアー**: Claude Code

---

## 目次

1. [概要](#概要)
2. [セキュリティ](#1-セキュリティ)
3. [パフォーマンス](#2-パフォーマンス)
4. [可読性](#3-可読性)
5. [保守性](#4-保守性)
6. [テスト](#5-テスト)
7. [TypeScript](#6-typescript)
8. [React/Vitestベストプラクティス](#7-reactvitestベストプラクティス)
9. [その他の発見事項](#8-その他の発見事項)
10. [まとめ](#まとめ)

---

## 概要

フロントエンドコードベースの詳細なレビューを実施しました。全体的に良好な実装が見られますが、特にWebSocket周りのメモリ管理と型アサーションの使用に改善の余地があります。

### ディレクトリ構造

```
frontend/src
├── features/           # Feature-basedな構造（良好）
│   ├── attachment/
│   ├── auth/
│   ├── bookmark/
│   ├── channel/
│   ├── dm/
│   ├── layout/
│   ├── link/
│   ├── member/
│   ├── message/
│   ├── notification/
│   ├── pin/
│   ├── reaction/
│   ├── search/
│   ├── settings/
│   ├── thread/
│   └── workspace/
├── lib/                # 共通ライブラリ
├── providers/          # グローバルProvider
├── routes/             # ルーティング定義
├── styles/             # スタイル
├── test/               # テストユーティリティ
└── types/              # 共通型定義
```

---

## 1. セキュリティ

### [must] XSS対策の確認

**状態**: ✅ **問題なし**

- `dangerouslySetInnerHTML`の使用は見つかりませんでした
- ユーザー入力は適切にエスケープされています

---

### [must] 認証トークンの安全な管理

**ファイル**: [src/lib/api/client.ts](../frontend/src/lib/api/client.ts)

**現在の実装** (行36-59):

```typescript
async function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  refreshPromise = (async () => {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return null;
    try {
      const { data, error } = await api.POST("/api/auth/refresh", {
        body: { refreshToken },
      });
      if (data && !error) {
        updateAuthTokens(data.accessToken, data.refreshToken);
        return data.accessToken;
      }
      return null;
    } catch {
      return null;  // ❌ エラーを握りつぶしている
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}
```

**問題点**:
- catch句でエラーを握りつぶすのではなく、ログ出力するか適切に処理すべき

**推奨修正**:

```typescript
} catch (error) {
  console.error('トークンのリフレッシュに失敗しました:', error);
  return null;
} finally {
```

---

### [recommend] WebSocketの認証トークン露出

**ファイル**: [src/lib/ws.ts:11-14](../frontend/src/lib/ws.ts#L11-L14)

**問題点**:

```typescript
function getWsUrl(token: string, workspaceId: string): string {
  const base = import.meta.env.VITE_WS_URL || "ws://localhost:8080";
  return `${base}/ws?token=${encodeURIComponent(token)}&workspaceId=${encodeURIComponent(workspaceId)}`;
}
```

**セキュリティリスク**:
- URLクエリパラメータでトークンを送信するのはセキュリティリスクがあります
- アクセスログやブラウザ履歴にトークンが残る可能性があります
- プロキシサーバーでトークンが露出する可能性があります

**推奨事項**:
1. WebSocket接続後の最初のメッセージとしてトークンを送信する方式に変更
2. または、Cookieベースの認証を使用

**修正例**:

```typescript
// 接続時
const ws = new WebSocket(`${base}/ws`);

// 接続確立後にトークンを送信
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: token,
    workspaceId: workspaceId
  }));
};
```

---

## 2. パフォーマンス

### [must] WsProviderでのメモリリーク懸念

**ファイル**: [src/providers/ws/WsProvider.tsx:22-45](../frontend/src/providers/ws/WsProvider.tsx#L22-L45)

**問題点**:

```typescript
useEffect(() => {
  if (!accessToken || !workspaceId) {
    setWsClient((prev) => {
      prev?.close();
      return null;
    });
    return;
  }
  if (!wsClient) {  // ❌ wsClientが依存配列にない
    setWsClient((prev) => {
      prev?.close();
      return null;
    });
    const instance = new WsClient(accessToken, workspaceId);
    setWsClient(instance);
  }

  return () => {
    setWsClient((prev) => {  // ❌ 毎回クリーンアップが実行される
      prev?.close();
      return null;
    });
  };
}, [accessToken, workspaceId]);  // ❌ wsClientが依存配列にない
```

**問題の詳細**:
1. `wsClient`が依存配列に含まれていないため、useEffectがwsClientの変更を検知できない
2. クリーンアップ関数が`accessToken`や`workspaceId`の変更のたびに実行される
3. 条件分岐内でのstate更新により、予期しない動作が発生する可能性

**修正案**:

```typescript
useEffect(() => {
  if (!accessToken || !workspaceId) {
    setWsClient((prev) => {
      prev?.close();
      return null;
    });
    return;
  }

  // 新しいインスタンスを作成
  const instance = new WsClient(accessToken, workspaceId);
  setWsClient(instance);

  // クリーンアップ時に作成したインスタンスのみをクローズ
  return () => {
    instance.close();
  };
}, [accessToken, workspaceId]);
```

---

### [must] MessagePanelでのメモリリーク

**ファイル**: [src/features/message/components/MessagePanel.tsx:43-60](../frontend/src/features/message/components/MessagePanel.tsx#L43-L60)

**問題点**:

```typescript
useEffect(() => {
  if (!wsClient || !currentChannelId) return;
  wsClient.joinChannel(currentChannelId);

  // new_message購読
  const handleNewMessage = (payload: NewMessagePayload) => {
    const result = messageWithThreadSchema.safeParse(payload.message);
    if (!result.success) return;
    setMessages((prev: MessageWithThread[]): MessageWithThread[] => {
      if (prev.some((m) => m.id === result.data.id)) return prev;
      return [...prev, result.data];
    });
  };

  wsClient.onNewMessage(handleNewMessage);  // ❌ 登録のみで解除していない

  return () => {
    wsClient.leaveChannel(currentChannelId);
    // ❌ ハンドラーのクリーンアップがない
  };
}, [wsClient, currentChannelId]);
```

**問題の詳細**:
- `wsClient.onNewMessage`で登録したハンドラーがクリーンアップされていない
- チャンネルを切り替えるたびに新しいハンドラーが追加され、古いハンドラーが残り続ける
- メモリリークとともに、同じイベントが複数回処理される可能性

**修正案**:

```typescript
useEffect(() => {
  if (!wsClient || !currentChannelId) return;
  wsClient.joinChannel(currentChannelId);

  const handleNewMessage = (payload: NewMessagePayload) => {
    const result = messageWithThreadSchema.safeParse(payload.message);
    if (!result.success) return;
    setMessages((prev: MessageWithThread[]): MessageWithThread[] => {
      if (prev.some((m) => m.id === result.data.id)) return prev;
      return [...prev, result.data];
    });
  };

  wsClient.onNewMessage(handleNewMessage);

  return () => {
    wsClient.offNewMessage(handleNewMessage);  // ✅ ハンドラーを解除
    wsClient.leaveChannel(currentChannelId);
  };
}, [wsClient, currentChannelId]);
```

**前提条件**: `WsClient`クラスに`offNewMessage`メソッドを追加する必要があります（後述）

---

### [recommend] 不要なレンダリングの最適化

**ファイル**: [src/features/message/components/MessagePanel.tsx:65-72](../frontend/src/features/message/components/MessagePanel.tsx#L65-L72)

**状態**: ✅ **良好**

```typescript
const dateTimeFormatter = useMemo(
  () =>
    new Intl.DateTimeFormat("ja-JP", {
      dateStyle: "short",
      timeStyle: "short",
    }),
  []
);
```

DateTimeFormatterの再作成を防ぐため`useMemo`を使用しており、適切です。

---

### [nits] BaseMessageInputの依存配列の問題

**ファイル**: [src/features/message/components/BaseMessageInput.tsx:49-73](../frontend/src/features/message/components/BaseMessageInput.tsx#L49-L73)

**問題点**:

```typescript
const handleBodyChange = useCallback(
  (newValue: string) => {
    setBody(newValue);
    // URLを検出してプレビューを追加
    const urlRegex = /https?:\/\/[^\s<>"{}|\\^`\[\]]+/g;
    const urls: string[] = newValue.match(urlRegex) || [];
    // ...
  },
  [previews, addPreview, removePreview]  // ❌ previewsを含めると再作成が頻繁に発生
);
```

**問題の詳細**:
- `previews`を依存配列に含めているため、プレビューが追加/削除されるたびに関数が再作成される
- これにより不要な再レンダリングが発生する可能性

**修正案**:

```typescript
const handleBodyChange = useCallback(
  (newValue: string) => {
    setBody(newValue);
    const urlRegex = /https?:\/\/[^\s<>"{}|\\^`\[\]]+/g;
    const urls: string[] = newValue.match(urlRegex) || [];

    // setPreviews内で最新のpreviewsを参照
    setPreviews((currentPreviews) => {
      // currentPreviewsを使用してロジックを実装
      // ...
    });
  },
  [addPreview, removePreview]  // ✅ previewsを削除
);
```

---

## 3. 可読性

### [recommend] マジックナンバーの使用

**ファイル**: [src/lib/ws.ts:18](../frontend/src/lib/ws.ts#L18)

**問題点**:

```typescript
private heartbeatInterval: number = 30000; // 30秒
```

**推奨事項**:

```typescript
// ファイル上部に定数定義
const WS_HEARTBEAT_INTERVAL = 30_000; // 30秒
const WS_RECONNECT_DELAY = 3_000; // 3秒

export class WsClient {
  private heartbeatInterval: number = WS_HEARTBEAT_INTERVAL;
  // ...
}
```

---

### [nits] console.logの使用

**該当ファイル**:
- [src/lib/ws.ts](../frontend/src/lib/ws.ts) (行126, 145)
- [src/features/workspace/hooks/useWorkspace.ts](../frontend/src/features/workspace/hooks/useWorkspace.ts) (行12, 30)
- その他多数

**問題点**:

```typescript
console.log("WebSocket接続成功");
console.warn("再接続試行中...");
console.error(error);
```

**推奨事項**:

開発環境と本番環境で適切にログを制御する仕組みを導入してください。

**実装例**:

```typescript
// src/lib/logger.ts
type LogLevel = 'debug' | 'info' | 'warn' | 'error';

class Logger {
  private isDevelopment = import.meta.env.DEV;

  debug(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.log(`[DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: unknown[]) {
    if (this.isDevelopment) {
      console.info(`[INFO] ${message}`, ...args);
    }
  }

  warn(message: string, ...args: unknown[]) {
    console.warn(`[WARN] ${message}`, ...args);
  }

  error(message: string, ...args: unknown[]) {
    console.error(`[ERROR] ${message}`, ...args);
  }
}

export const logger = new Logger();
```

**使用例**:

```typescript
import { logger } from '@/lib/logger';

logger.info('WebSocket接続成功');
logger.error('接続エラー:', error);
```

---

## 4. 保守性

### [must] 型アサーションの不適切な使用

CLAUDE.mdのガイドライン:
> 修正にあたり、any/unknown などの型を使用することや、型アサーション・型ガードを使用することを禁止します。その実装にふさわしい型を書くか、ライブラリから提供されているものをインポートして使用してください。

以下のファイルで不必要な型アサーションが見つかりました。

---

#### 4.1 useParticipatingThreads

**ファイル**: [src/features/thread/hooks/useParticipatingThreads.ts](../frontend/src/features/thread/hooks/useParticipatingThreads.ts)

**問題点** (行31, 54):

```typescript
return { items: [], next_cursor: undefined } as unknown as ParticipatingThreadsOutput;
// ...
return parsed.data as unknown as ParticipatingThreadsOutput;
```

**修正案**:

```typescript
// スキーマの出力型を使用
import type { components } from '@/lib/api/schema';

type ParticipatingThreadsOutput = components['schemas']['ParticipatingThreadsOutput'];

// 型アサーションを削除
return { items: [], next_cursor: undefined }; // 型が合わない場合はスキーマ定義を確認
```

---

#### 4.2 usePinnedMessages

**ファイル**: [src/features/pin/hooks/usePinnedMessages.ts:40,50](../frontend/src/features/pin/hooks/usePinnedMessages.ts#L40)

**問題点**:

```typescript
if (channelId === null) return { pins: [], nextCursor: null } as PinnedListResponse;
// ...
return data as unknown as PinnedListResponse;
```

**修正案**:

```typescript
// APIスキーマから正しい型をimport
import type { components } from '@/lib/api/schema';

type PinnedListResponse = components['schemas']['PinnedListResponse'];

// 型アサーションを削除し、型定義を修正
if (channelId === null) {
  return { pins: [], nextCursor: null }; // 型が合わない場合はPinnedListResponseの定義を確認
}

// data の型は既に正しいはずなので、アサーション不要
return data;
```

---

#### 4.3 useDM

**ファイル**: [src/features/dm/hooks/useDM.ts:24,46](../frontend/src/features/dm/hooks/useDM.ts#L24)

**問題点**:

```typescript
return response.data as DMOutput[];
// ...
return response.data as DMOutput;
```

**修正案**:

`response.data`の型は既に正しいはずです。型アサーションを削除してください。

```typescript
// APIクライアントの型定義が正しければ、アサーション不要
return response.data;
```

型エラーが発生する場合は、`DMOutput`の型定義とAPIスキーマの定義を確認してください。

---

#### 4.4 useLogin / useRegister

**ファイル**:
- [src/features/auth/hooks/useLogin.ts:23](../frontend/src/features/auth/hooks/useLogin.ts#L23)
- [src/features/auth/hooks/useRegister.ts:23](../frontend/src/features/auth/hooks/useRegister.ts#L23)

**問題点**:

```typescript
return response as AuthResponse;
```

**修正案**:

```typescript
// responseは既に正しい型を持っているはずなので、アサーション削除
return response;
```

型エラーが発生する場合は、関数の戻り値の型定義を見直してください。

---

### [recommend] window オブジェクトの直接使用

**ファイル**: [src/providers/store/auth.ts:64-81](../frontend/src/providers/store/auth.ts#L64-L81)

**現在の実装**:

```typescript
if (typeof window !== "undefined") {
  const legacyAccessToken = window.localStorage.getItem("accessToken");
  const legacyRefreshToken = window.localStorage.getItem("refreshToken");
  // ...
  window.localStorage.removeItem("accessToken");
  window.localStorage.removeItem("refreshToken");
}
```

**状態**: 🟡 **改善の余地あり**

**良い点**:
- SSR対応のための`typeof window !== "undefined"`チェックは適切

**推奨事項**:

ストレージアクセスを抽象化したユーティリティ関数を作成することで、テストしやすくなります。

**実装例**:

```typescript
// src/lib/storage.ts
export const storage = {
  getItem: (key: string): string | null => {
    if (typeof window === "undefined") return null;
    return window.localStorage.getItem(key);
  },

  setItem: (key: string, value: string): void => {
    if (typeof window === "undefined") return;
    window.localStorage.setItem(key, value);
  },

  removeItem: (key: string): void => {
    if (typeof window === "undefined") return;
    window.localStorage.removeItem(key);
  },
};

// 使用例
const legacyAccessToken = storage.getItem("accessToken");
storage.removeItem("accessToken");
```

---

### [recommend] エラーハンドリングの一貫性

**ファイル**: [src/features/workspace/hooks/useWorkspace.ts:5-18](../frontend/src/features/workspace/hooks/useWorkspace.ts#L5-L18)

**問題点**:

```typescript
export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces", {});

      if (error || !data) {
        console.error(error);
        return [];  // ❌ エラーを隠蔽している
      }

      return data.workspaces;
    },
  });
}
```

**問題の詳細**:
- エラー時に空配列を返すのは、エラーの隠蔽につながる
- ユーザーにエラーが発生したことが伝わらない
- 他のフックでは`throw new Error()`を使用しているため、一貫性がない

**修正案**:

```typescript
export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data, error } = await api.GET("/api/workspaces", {});

      if (error || !data) {
        throw new Error(error?.error ?? "ワークスペースの取得に失敗しました");
      }

      return data.workspaces;
    },
  });
}
```

React Queryが自動的にエラー状態を管理し、UIでエラー表示が可能になります。

---

## 5. テスト

### [recommend] テストカバレッジの不足

以下の重要なファイルにテストが不足しています:

| ファイル | 重要度 | 理由 |
|---------|--------|------|
| [src/lib/ws.ts](../frontend/src/lib/ws.ts) | 🔴 高 | WebSocketクライアントの中核ロジック |
| [src/providers/ws/WsProvider.tsx](../frontend/src/providers/ws/WsProvider.tsx) | 🔴 高 | アプリケーション全体のWebSocket管理 |
| [src/features/message/components/MessagePanel.tsx](../frontend/src/features/message/components/MessagePanel.tsx) | 🟡 中 | メッセージ表示の主要コンポーネント |
| [src/features/message/components/BaseMessageInput.tsx](../frontend/src/features/message/components/BaseMessageInput.tsx) | 🟡 中 | メッセージ入力の主要コンポーネント |
| [src/features/attachment/hooks/useFileUpload.ts](../frontend/src/features/attachment/hooks/useFileUpload.ts) | 🟡 中 | ファイルアップロードロジック |

**推奨事項**:

優先度の高いものから順次テストを追加してください。

**テスト例 (ws.ts)**:

```typescript
// src/lib/ws.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { WsClient } from './ws';

describe('WsClient', () => {
  let client: WsClient;
  const mockToken = 'test-token';
  const mockWorkspaceId = 'workspace-123';

  beforeEach(() => {
    // WebSocketのモック
    global.WebSocket = vi.fn(() => ({
      send: vi.fn(),
      close: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })) as any;
  });

  afterEach(() => {
    client?.close();
  });

  it('正常に接続できること', () => {
    client = new WsClient(mockToken, mockWorkspaceId);
    expect(global.WebSocket).toHaveBeenCalledWith(
      expect.stringContaining('/ws')
    );
  });

  it('メッセージハンドラーを登録できること', () => {
    client = new WsClient(mockToken, mockWorkspaceId);
    const handler = vi.fn();
    client.onNewMessage(handler);
    // ハンドラーが登録されたことを検証
  });

  // 他のテストケース...
});
```

---

### [nits] テストでの型アサーション

**ファイル**: [src/features/channel/hooks/useChannelMembers.test.ts](../frontend/src/features/channel/hooks/useChannelMembers.test.ts)

**問題点** (行71, 91, 111):

```typescript
response: {} as Response,
```

**推奨事項**:

モックレスポンスオブジェクトを適切に定義するか、`vi.fn()`を使用してください。

**修正例**:

```typescript
// モックResponseの作成
const createMockResponse = (data: unknown): Response => ({
  ok: true,
  status: 200,
  statusText: 'OK',
  json: async () => data,
  text: async () => JSON.stringify(data),
  // 他の必要なプロパティ...
} as Response);

// 使用例
response: createMockResponse({ members: [] }),
```

---

## 6. TypeScript

### [must] interfaceの使用

**ファイル**: [src/test/vitest.d.ts:6,8](../frontend/src/test/vitest.d.ts#L6-L8)

**問題点**:

```typescript
interface Assertion<T = any> extends TestingLibraryMatchers<T, void> {}
interface AsymmetricMatchersContaining extends TestingLibraryMatchers<any, void> {}
```

**ガイドライン違反**:

CLAUDE.mdで「型定義に`interface`を使用せず、必ず`type`を使用してください」と指定されています。

**考察**:

このファイルはVitestの型定義の拡張であり、元の定義が`interface`を使用しているため、`interface`での拡張が技術的に必要な可能性があります。

**推奨対応**:

1. 可能であれば`type`に変更:

```typescript
type Assertion<T = any> = TestingLibraryMatchers<T, void>;
type AsymmetricMatchersContaining = TestingLibraryMatchers<any, void>;
```

2. 技術的に不可能な場合は、このファイルをガイドラインの例外として明記

---

### [recommend] anyの使用

**ファイル**: [src/routes/routeTree.gen.ts](../frontend/src/routes/routeTree.gen.ts)

**状態**: ✅ **問題なし（自動生成ファイル）**

このファイルは自動生成されたファイルであり、複数の`as any`が含まれていますが、ヘッダーコメントに「変更禁止」が明記されています。

```typescript
/* prettier-ignore-start */

/* eslint-disable */

// @ts-nocheck

// noinspection JSUnusedGlobalSymbols
```

---

### [recommend] unknownの適切な使用

**ファイル**: [src/types/wsEvents.ts:27-32](../frontend/src/types/wsEvents.ts#L27-L32)

**問題点**:

```typescript
export type NewMessagePayload = { channel_id: string; message: Record<string, unknown> };
export type MessageUpdatedPayload = { channel_id: string; message: Record<string, unknown> };
export type MessageDeletedPayload = { channel_id: string; deleteData: Record<string, unknown> };
```

**推奨事項**:

`Record<string, unknown>`ではなく、適切な型定義を使用してください。

**修正案**:

```typescript
import type { MessageWithThread } from '@/features/message/types';

export type NewMessagePayload = {
  channel_id: string;
  message: MessageWithThread;
};

export type MessageUpdatedPayload = {
  channel_id: string;
  message: MessageWithThread;
};

export type MessageDeletedPayload = {
  channel_id: string;
  deleteData: {
    id: string;
    deleted_at: string;
  };
};
```

---

## 7. React/Vitestベストプラクティス

### [recommend] useEffectの依存配列

**ファイル**: [src/features/message/components/BaseMessageInput.tsx:108-114](../frontend/src/features/message/components/BaseMessageInput.tsx#L108-L114)

**問題点**:

```typescript
useEffect(() => {
  if (onReset) {
    setBody("");
    clearPreviews();
    clearAttachments();
  }
}, [onReset, clearPreviews, clearAttachments]);
```

**問題の詳細**:
- `onReset`は関数プロパティであり、依存配列に含めると親コンポーネントの再レンダリングのたびに実行される可能性がある
- 設計的に`onReset`が呼び出しトリガーではなく、外部からのリセット命令として機能するなら、別のアプローチを検討すべき

**推奨修正**:

パターン1: `onReset`を数値カウンターに変更

```typescript
// 親コンポーネント
const [resetCounter, setResetCounter] = useState(0);
const handleReset = () => setResetCounter(prev => prev + 1);

<BaseMessageInput resetTrigger={resetCounter} />

// BaseMessageInput
useEffect(() => {
  if (resetTrigger > 0) {
    setBody("");
    clearPreviews();
    clearAttachments();
  }
}, [resetTrigger, clearPreviews, clearAttachments]);
```

パターン2: `useImperativeHandle`を使用

```typescript
// BaseMessageInput
const BaseMessageInput = forwardRef((props, ref) => {
  useImperativeHandle(ref, () => ({
    reset: () => {
      setBody("");
      clearPreviews();
      clearAttachments();
    }
  }));
  // ...
});

// 親コンポーネント
const inputRef = useRef<{ reset: () => void }>(null);
const handleReset = () => inputRef.current?.reset();
```

---

### [nits] Hooksのルール違反の可能性

**ファイル**: [src/features/channel/components/ChannelList.tsx:31-38](../frontend/src/features/channel/components/ChannelList.tsx#L31-L38)

**コード**:

```typescript
useEffect(() => {
  if (channels && channels.length > 0 && currentChannelId === null) {
    const firstChannel = channels[0];
    if (firstChannel) {
      setCurrentChannel(firstChannel.id);
    }
  }
}, [channels, currentChannelId, setCurrentChannel]);
```

**状態**: ✅ **問題なし**

条件付きでstate更新していますが、useEffect内なのでHooksのルールには違反していません。

---

### [recommend] 未使用の引数

**ファイル**: [src/features/message/components/BaseMessageInput.tsx](../frontend/src/features/message/components/BaseMessageInput.tsx)

**問題点**:

```typescript
type Props = {
  channelId?: string;  // ❌ オプショナルだが、useFileUploadでは必須として使用
  // ...
};

// 使用箇所
const { uploadFiles } = useFileUpload(channelId!);  // ❌ 非nullアサーション
```

**推奨修正**:

```typescript
type Props = {
  channelId: string;  // ✅ 必須に変更
  // ...
};

const { uploadFiles } = useFileUpload(channelId);  // ✅ アサーション不要
```

---

## 8. その他の発見事項

### [must] WsClientのイベントハンドラー解除機構の欠如

**ファイル**: [src/lib/ws.ts](../frontend/src/lib/ws.ts)

**問題点**:
- `onNewMessage`などのイベントハンドラーを登録するメソッドはあるが、解除するメソッドがない
- これにより、コンポーネントのアンマウント時にハンドラーがクリーンアップされず、メモリリークが発生

**現在の実装**:

```typescript
export class WsClient {
  private handlers: {
    new_message: ((payload: WsEventPayloadMap["new_message"]) => void)[];
    message_updated: ((payload: WsEventPayloadMap["message_updated"]) => void)[];
    // ...
  };

  public onNewMessage(cb: (payload: WsEventPayloadMap["new_message"]) => void) {
    this.handlers.new_message.push(cb);
  }

  // ❌ offNewMessage メソッドが存在しない
}
```

**修正案**:

```typescript
export class WsClient {
  // ... 既存のコード ...

  // ハンドラー解除メソッドを追加
  public offNewMessage(cb: (payload: WsEventPayloadMap["new_message"]) => void) {
    const index = this.handlers.new_message.indexOf(cb);
    if (index > -1) {
      this.handlers.new_message.splice(index, 1);
    }
  }

  public offMessageUpdated(cb: (payload: WsEventPayloadMap["message_updated"]) => void) {
    const index = this.handlers.message_updated.indexOf(cb);
    if (index > -1) {
      this.handlers.message_updated.splice(index, 1);
    }
  }

  public offMessageDeleted(cb: (payload: WsEventPayloadMap["message_deleted"]) => void) {
    const index = this.handlers.message_deleted.indexOf(cb);
    if (index > -1) {
      this.handlers.message_deleted.splice(index, 1);
    }
  }

  // 他のイベントタイプにも同様のメソッドを追加
  // offReactionAdded, offReactionRemoved, offChannelUpdated, etc.

  // すべてのハンドラーをクリアするメソッド（オプション）
  public clearAllHandlers() {
    this.handlers = {
      new_message: [],
      message_updated: [],
      message_deleted: [],
      reaction_added: [],
      reaction_removed: [],
      channel_updated: [],
      member_joined: [],
      member_left: [],
    };
  }
}
```

**使用例**:

```typescript
// コンポーネント内
useEffect(() => {
  if (!wsClient) return;

  const handleNewMessage = (payload: NewMessagePayload) => {
    // 処理...
  };

  wsClient.onNewMessage(handleNewMessage);

  return () => {
    wsClient.offNewMessage(handleNewMessage);  // ✅ クリーンアップ
  };
}, [wsClient]);
```

---

### [nits] URLの構築

**ファイル**: 複数のファイル

**問題点**:

```typescript
const url = `${window.location.origin}/app/${currentWorkspaceId}/${currentChannelId}?message=${messageId}`;
```

**推奨事項**:

Tanstack RouterのAPIを使用してURLを生成することで、型安全性が向上します。

**修正例**:

```typescript
import { useRouter } from '@tanstack/react-router';

const router = useRouter();

// 型安全なURL生成
const url = router.buildLocation({
  to: '/app/$workspaceId/$channelId',
  params: {
    workspaceId: currentWorkspaceId,
    channelId: currentChannelId,
  },
  search: {
    message: messageId,
  },
}).href;
```

---

## まとめ

### 🔴 優先度：高 [must] - 即座に対応すべき項目

| # | 項目 | ファイル | 影響 |
|---|------|---------|------|
| 1 | WsProviderのメモリリーク修正 | [WsProvider.tsx:22-45](../frontend/src/providers/ws/WsProvider.tsx#L22-L45) | 無限ループとメモリリークの可能性 |
| 2 | MessagePanelのWebSocketハンドラークリーンアップ | [MessagePanel.tsx:43-60](../frontend/src/features/message/components/MessagePanel.tsx#L43-L60) | メモリリーク |
| 3 | WsClientへのハンドラー解除機能の追加 | [ws.ts](../frontend/src/lib/ws.ts) | メモリリーク対策 |
| 4 | 型アサーションの削除 | 複数ファイル | 型安全性の向上 |

---

### 🟡 優先度：中 [recommend] - 計画的に対応すべき項目

| # | 項目 | ファイル | 理由 |
|---|------|---------|------|
| 1 | WebSocketのトークン送信方法の見直し | [ws.ts:11-14](../frontend/src/lib/ws.ts#L11-L14) | セキュリティリスク |
| 2 | エラーハンドリングの一貫性向上 | [useWorkspace.ts:5-18](../frontend/src/features/workspace/hooks/useWorkspace.ts#L5-L18) | UX改善 |
| 3 | テストカバレッジの向上 | 複数ファイル | 品質保証 |
| 4 | console.logの適切な管理 | 複数ファイル | 本番環境での不要なログ出力 |
| 5 | windowオブジェクトアクセスの抽象化 | [auth.ts:64-81](../frontend/src/providers/store/auth.ts#L64-L81) | テスタビリティ向上 |
| 6 | unknownの適切な型定義 | [wsEvents.ts:27-32](../frontend/src/types/wsEvents.ts#L27-L32) | 型安全性 |
| 7 | useEffectの依存配列最適化 | [BaseMessageInput.tsx:108-114](../frontend/src/features/message/components/BaseMessageInput.tsx#L108-L114) | パフォーマンス |
| 8 | 認証エラーハンドリングの改善 | [client.ts:36-59](../frontend/src/lib/api/client.ts#L36-L59) | デバッグ容易性 |

---

### 🔵 優先度：低 [nits] - 時間があれば対応

| # | 項目 | ファイル |
|---|------|---------|
| 1 | マジックナンバーの定数化 | [ws.ts:18](../frontend/src/lib/ws.ts#L18) |
| 2 | interfaceからtypeへの変更 | [vitest.d.ts:6,8](../frontend/src/test/vitest.d.ts#L6-L8) |
| 3 | URL構築の型安全化 | 複数ファイル |
| 4 | テストでの型アサーション改善 | [useChannelMembers.test.ts](../frontend/src/features/channel/hooks/useChannelMembers.test.ts) |

---

### ✅ 評価できる点

1. **Feature-based なディレクトリ構造** - コードの保守性が高い
2. **XSS対策** - `dangerouslySetInnerHTML`を使用していない
3. **パフォーマンス最適化** - `useMemo`、`useCallback`の適切な使用
4. **SSR対応** - `typeof window !== "undefined"`チェック
5. **テスト文化** - 主要コンポーネントにテストが存在
6. **型安全性** - TypeScriptを活用した実装

---

## 推奨対応順序

### フェーズ1: 緊急対応（1-2日）

1. WsClientにハンドラー解除メソッドを追加
2. WsProviderのuseEffectを修正
3. MessagePanelのハンドラークリーンアップを追加

### フェーズ2: 重要な改善（1週間）

4. 型アサーションの削除（複数ファイル）
5. WebSocketのトークン送信方法の変更
6. エラーハンドリングの統一

### フェーズ3: 品質向上（2週間）

7. テストカバレッジの向上
8. ロガーライブラリの導入
9. 型定義の改善（wsEvents.ts）

### フェーズ4: リファクタリング（随時）

10. コード可読性の向上（マジックナンバー、URL構築など）
11. パフォーマンス最適化の微調整

---

## 総評

フロントエンドコードベースは、全体的に**良好な品質**を保っています。特にFeature-basedな構造、型安全性への配慮、テストの存在など、多くの良い実践が見られます。

しかし、**WebSocket周りのメモリ管理**と**型アサーションの過剰使用**には注意が必要です。これらは本番環境でのパフォーマンス低下や予期しないバグにつながる可能性があります。

優先度の高い項目から順次対応することで、より堅牢で保守性の高いコードベースに改善できます。

---

**レビュー実施**: Claude Code
**最終更新**: 2025-10-30
