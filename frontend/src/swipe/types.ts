export type Card = {
  id: string;
  title: string;
  description: string;
};

export type SwipeDirection = "left" | "right";

export type VoteEvent = {
  roomId: string;
  cardId: string;
  direction: SwipeDirection;
  ts: number;
};
