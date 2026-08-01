import { API_BASE_URL } from "./client";

export function getProxyImageUrl(photoName?: string): string | null {
  if (!photoName) return null;
  return `${API_BASE_URL}/image?name=${encodeURIComponent(photoName)}`;
}
