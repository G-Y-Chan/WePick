import { api } from "./client";

export type Message = {
  Header: string;
  Body: string;
};

export function postEmailAddress(email: string) {
  return api.post<Message>("/post-email", email);
}