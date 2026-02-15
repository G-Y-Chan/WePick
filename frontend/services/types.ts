export type Message = {
  Header: string;
  Body?: string;
  VoteObj?: Vote
};

type Vote = {
  Id: string;
  Result: string; //"ACCEPT" or "REJECT"
}