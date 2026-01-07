import { api } from "./client";

export type Message = {
  Header: string;
  Body: string;
};

export async function getRoomCode(): Promise<string> {
  const res = await api.get<Message>("/get-room-code");
  return res.Body;
}

export async function verifyRoomCode(roomCode: number): Promise<string> {
  const res = await api.post<Message>("/verify-room-code", roomCode.toString());
  return res.Body;
}