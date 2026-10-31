const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export type User = {
  id: number;
  email: string;
  username: string;
  full_name: string;
  avatar_url: string;
  role: string;
  created_at: string;
  updated_at: string;
};

export type AuthResponse = {
  token: string;
  user: User;
};

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  // Cookie httpOnly dikirim otomatis via credentials:include.
  // Tidak lagi baca localStorage / header Authorization.
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
    credentials: "include",
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error((data as { error?: string }).error || res.statusText);
  }
  return data as T;
}

export const api = {
  register: (body: { email: string; username: string; password: string; full_name?: string }) =>
    request<AuthResponse>("/api/auth/register", { method: "POST", body: JSON.stringify(body) }),
  login: (body: { email: string; password: string }) =>
    request<AuthResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(body) }),
  logout: () => request<{ message: string }>("/api/auth/logout", { method: "POST" }),
  me: () => request<{ user: User }>("/api/auth/me"),
  listUsers: () => request<{ users: User[]; total: number }>("/api/users"),
  getUser: (id: number) => request<{ user: User }>(`/api/users/${id}`),
  updateProfile: (body: { full_name?: string; avatar_url?: string; username?: string }) =>
    request<{ user: User }>("/api/users/me", { method: "PATCH", body: JSON.stringify(body) }),
};
