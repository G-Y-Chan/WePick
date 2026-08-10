import { API_BASE_URL } from "./client";

export function getImageUrl(photoUrl?: string | null, photoName?: string | null): string | null {
  if (photoUrl) return photoUrl;
  if (!photoName) return null;
  return `${API_BASE_URL}/image/${encodeURIComponent(photoName)}`;
}

export function getProxyImageUrl(photoName?: string): string | null {
  if (!photoName) return null;
  return `${API_BASE_URL}/image/${encodeURIComponent(photoName)}`;
}
