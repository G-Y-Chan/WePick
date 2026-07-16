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

// Updated to include the host's GPS coordinates
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
  
  // Bundle the roomCode and filters into a single JSON payload
  const payload = {
    roomCode,
    filters,
  };

  // Send the payload to your backend
  const res = await api.post<Message>("/start-room", payload);
  return res.Body?.toLowerCase();
}

// Updated to accept roomCode, use a GET request, and pass the code as a query parameter
export async function getCardData(roomCode: string): Promise<Card[] | undefined> {
  const res = await api.get<Message>(`/get-card-data?code=${roomCode}`);
  return res.Cards;
}