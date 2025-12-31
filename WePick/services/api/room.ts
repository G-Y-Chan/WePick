import { api } from "./client";

export type Message = {
  Header: string;
  Body: string;
};

export async function getRoomCode() {
  const res = await api.get<Message>("/get-room-code");
  return res.Body;
}