import { useQuery } from "@connectrpc/connect-query";

import { browserTimeZone } from "#/features/insights/utils/chart";
import { InsightService } from "#/gen/chat/v1/insight_service_pb";

export const useInsights = (workspaceId: string) =>
  useQuery(InsightService.method.getInsights, { timeZone: browserTimeZone(), workspaceId });
