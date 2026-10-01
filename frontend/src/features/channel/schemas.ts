import { z } from "zod";

import { BrowsableChannelMembership, BrowsableChannelSort } from "#/gen/chat/v1/channel_service_pb";

export const browseMembershipValues = ["all", "joined", "notJoined"] as const;
type BrowseMembership = (typeof browseMembershipValues)[number];

export const browseMembershipMessages: Record<BrowseMembership, BrowsableChannelMembership> = {
  all: BrowsableChannelMembership.UNSPECIFIED,
  joined: BrowsableChannelMembership.JOINED,
  notJoined: BrowsableChannelMembership.NOT_JOINED,
};

export const browseSortValues = ["name", "members"] as const;
type BrowseSort = (typeof browseSortValues)[number];

export const browseSortMessages: Record<BrowseSort, BrowsableChannelSort> = {
  members: BrowsableChannelSort.MEMBER_COUNT,
  name: BrowsableChannelSort.UNSPECIFIED,
};

export const browseChannelsSearchSchema = z.object({
  membership: z.enum(browseMembershipValues).default("all").catch("all"),
  page: z.number().int().min(1).default(1).catch(1),
  q: z.string().default("").catch(""),
  sort: z.enum(browseSortValues).default("name").catch("name"),
});
