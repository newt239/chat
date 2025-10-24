export type LinkInfo = {
  id: string;
  url: string;
  title?: string | null;
  description?: string | null;
  imageUrl?: string | null;
  siteName?: string | null;
  cardType?: string | null;
};

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
