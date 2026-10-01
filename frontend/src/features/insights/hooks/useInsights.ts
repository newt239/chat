import { useQuery } from "@connectrpc/connect-query";

import { InsightService } from "#/gen/chat/v1/insight_service_pb";
import { useDateFormat } from "#/hooks/useDateFormat";

// 日ごと・時間帯ごとの集計はプロフィールのタイムゾーンで区切る
export const useInsights = (workspaceId: string) => {
  const { timeZone } = useDateFormat();
  return useQuery(InsightService.method.getInsights, { timeZone, workspaceId });
};
