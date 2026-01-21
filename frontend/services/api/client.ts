// Go server endpoint
const API_BASE_URL = "http://localhost:8090";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  try {
    const res = await fetch(`${API_BASE_URL}${path}`, {
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
      ...options,
    });

    const text = await res.text(); // Read raw text first

    if (!res.ok) {
      // HTTP error (4xx / 5xx)
      let errorMessage = `Request failed with status ${res.status}`;
      try {
        const errorData = JSON.parse(text);
        errorMessage = errorData.Body || errorData.Message || errorMessage;
      } catch (e) {
        errorMessage = text || errorMessage;
      }
      throw new Error(errorMessage);
    }

    if (!text) {
      // Empty body → return null or {} depending on your needs
      return null as T;
    }

    try {
      return JSON.parse(text);
    } catch (e) {
      console.error("Invalid JSON response:", text);
      throw e;
    }
  } catch (err: unknown) {
    if (err instanceof Error) {
      console.log("Network error:", err.message);
    }
    return null as T
  }
}


export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, {
      method: "POST",
      body: JSON.stringify(body),
    }),
};
