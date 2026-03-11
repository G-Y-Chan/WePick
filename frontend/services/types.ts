import { Card } from "@/src/swipe/types";

export type Message = {
  Header: string;
  Body?: string;
  VoteObj?: Vote
  Cards?: Card[]
};

type Vote = {
  Id: string;
  Result: string; //"ACCEPT" or "REJECT"
}