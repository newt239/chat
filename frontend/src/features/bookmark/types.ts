export type Bookmark = {
  userId: string;
  messageId: string;
  createdAt: string;
};

export type BookmarkWithMessage = {
  userId: string;
  message: {
    id: string;
    channelId: string;
    userId: string;
    parentId?: string;
    body: string;
    createdAt: string;
    editedAt?: string;
    deletedAt?: string;
  };
  createdAt: string;
};
