import type { Messages } from "../../messages";

export const insights: Messages["insights"] = {
  charts: {
    channels: {
      note: "Last 30 days · Excluding DMs",
      title: "Messages by channel",
    },
    daily: {
      note: "Last 30 days · Today is partial",
      title: "Messages per day",
    },
    heatmap: {
      cell: "{{weekday}} {{hour}}:00 · Average {{value}}",
      hint: "Hover over a cell to see its count",
      less: "Less",
      more: "More",
      note: "Weekday × hour (average of the last {{weeks}} weeks)",
      title: "Busiest times",
    },
    mine: {
      note: "Last 7 days · Only visible to you",
      title: "Your messages",
    },
    popular: {
      note: "Messages in the last 7 days",
      title: "Popular channels",
    },
    storage: {
      categories: {
        audio: "Audio",
        file: "Files",
        image: "Images",
        unspecified: "Other",
        video: "Videos",
      },
      note: "Total {{total}}",
      title: "Storage breakdown",
    },
  },
  empty: "No data yet",
  kpi: {
    activeMembers: "Active members",
    activeRate: "Active rate",
    messages: "Messages (30 days)",
    myMessages: "Your messages (7 days)",
    ofMembers: "/ {{count}}",
    percentDelta: "{{value}}% vs previous",
    pointDelta: "{{value}} pt vs previous",
    storage: "Storage",
    storageDelta: "{{value}} vs previous",
    unitCount: "",
    unitPeople: "",
  },
  loadFailed: "Couldn't load insights",
  note: "Per-member breakdowns and the audit log are in the admin console.",
  openAdmin: "Open admin",
  table: {
    active: "Active",
    category: "Type",
    channel: "Channel",
    count: "Count",
    date: "Date",
    files: "Files",
    showChart: "Show chart",
    showTable: "Show table",
    size: "Size",
    weekday: "Weekday",
  },
  title: "Insights",
  tooltip: {
    count: "{{label}} · {{value}}",
    partial: " (partial)",
  },
};
