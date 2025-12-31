import { api } from "./client";

export type Message = {
  Header: string;
  Body: string;
};

export function fetchTestMessage() {
  return api.get<Message>("/test");
}
