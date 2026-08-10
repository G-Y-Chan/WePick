export type Card = {
  id: string;
  title: string;
  category: string;
  priceLevel: string;
  rating: number;
  reviewCount: number;
  openNow: boolean;
  summary: string;
  address: string;
  photoName?: string;
  photoUrl?: string;
};

export type SwipeDirection = "left" | "right";

export type VoteEvent = {
  roomId: string;
  cardId: string;
  direction: SwipeDirection;
  ts: number;
};