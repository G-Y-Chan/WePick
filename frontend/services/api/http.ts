import { api } from "./client";
import { Message } from "../types"
import { Card } from "@/src/swipe/types";

export async function getRoomCode(): Promise<string | undefined> {
  const res = await api.get<Message>("/rooms");
  return res.Body;
}

export async function joinRoom(roomCode: string | number): Promise<string | undefined> {
  const res = await api.post<Message>(`/rooms/${roomCode}/join`);
  return res.Body?.toLowerCase();
}

export interface RoomFilters {
  latitude: number;
  longitude: number;
  radius: number;
  maxPrice: number;
  category: string;
  openNow: boolean;
}

export async function startRoom(roomCode: string, filters: RoomFilters): Promise<string | undefined> {
  console.log(`Attempting to start room: ${roomCode} with filters:`, filters);
  
  const payload = {
    filters,
  };

  const res = await api.post<Message>(`/rooms/${roomCode}/start`, payload);
  return res.Body?.toLowerCase();
}

export async function getCardData(roomCode: string): Promise<Card[] | undefined> {
  const res = await api.get<Message>(`/rooms/${roomCode}/cards`);
  return res.Cards;
}