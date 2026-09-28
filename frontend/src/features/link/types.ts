export type OGPData = {
  title?: string;
  description?: string;
  imageUrl?: string;
  siteName?: string;
  cardType?: string;
};

export type LinkPreview = {
  url: string;
  ogpData: OGPData;
  isLoading: boolean;
  error?: string;
};
