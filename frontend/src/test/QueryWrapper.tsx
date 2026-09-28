import type { ReactNode } from "react";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// API を呼ばない前提なのでクライアントは共有する
const client = new QueryClient();

// クエリのフック（表示名の解決など）を使うが API は呼ばない部品を単体で描画するためのラッパー
export const QueryWrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={client}>{children}</QueryClientProvider>
);
