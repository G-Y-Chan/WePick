import { api } from "./client";

export type Message = {
  Header: string;
  Body: string;
};

export async function getRoomCode(): Promise<string> {
  const res = await api.get<Message>("/get-room-code");
  return res.Body;
}

export async function joinRoom(roomCode: number): Promise<string> {
  const res = await api.post<Message>("/join-room", roomCode.toString());
  return res.Body;
}

export async function startRoom(roomCode: string): Promise<string> {
  const res = await api.post<Message>("/start-room", roomCode);
  return res.Body;
}