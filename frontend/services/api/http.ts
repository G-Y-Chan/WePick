import { api } from "./client";
import { Message } from "../types"
import { Card } from "@/src/swipe/types";

export async function getRoomCode(): Promise<string | undefined> {
  const res = await api.get<Message>("/get-room-code");
  return res.Body;
}

export async function joinRoom(roomCode: number): Promise<string | undefined> {
  const res = await api.post<Message>("/join-room", roomCode.toString());
  return res.Body?.toLowerCase();
}

export async function startRoom(roomCode: string): Promise<string | undefined> {
  console.log(`Attempting to start room: ${roomCode}`);
  const res = await api.post<Message>("/start-room", roomCode);
  return res.Body?.toLowerCase();
}

export async function getCardData(location: string): Promise<Card[] | undefined> {
  const res = await api.post<Message>("/get-card-data", location);
  return res.Cards;
}