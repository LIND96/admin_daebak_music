const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export type User = {
  id: number;
  email: string;
  username: string;
  full_name: string;
  avatar_url: string;
  role: string;
  auth_provider?: string;
  created_at: string;
  updated_at: string;
};

export type SessionResponse = {
  access_token: string;
  refresh_token: string;
  /** Alias for access_token, kept for backward compat. */
  token: string;
  expires_in: number;
  refresh_expires_in: number;
  user: User;
};

// Backward compat: existing code reads .token / .user.
export type AuthResponse = SessionResponse;

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  // Cookie httpOnly dikirim otomatis via credentials:include.
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
  // Register publik dimatikan. Admin dibuat via bootstrap / POST /api/users oleh admin.
  bootstrapAdmin: (body: { email: string; username: string; password: string; full_name?: string; setup_key: string }) =>
    request<SessionResponse>("/api/auth/bootstrap-admin", { method: "POST", body: JSON.stringify(body) }),
  login: (body: { email: string; password: string }) =>
    request<SessionResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(body) }),
  googleLogin: (body: { id_token: string; device_info?: string }) =>
    request<SessionResponse>("/api/auth/google", { method: "POST", body: JSON.stringify(body) }),
  refresh: (body?: { refresh_token?: string }) =>
    request<SessionResponse>("/api/auth/refresh", { method: "POST", body: JSON.stringify(body ?? {}) }),
  logout: (body?: { refresh_token?: string }) =>
    request<{ message: string }>("/api/auth/logout", { method: "POST", body: JSON.stringify(body ?? {}) }),
  me: () => request<{ user: User }>("/api/auth/me"),
  listUsers: () => request<{ users: User[]; total: number }>("/api/users"),
  getUser: (id: number) => request<{ user: User }>(`/api/users/${id}`),
  createUser: (body: { email: string; username: string; password: string; full_name?: string; role: "admin" | "customer" }) =>
    request<{ user: User }>("/api/users", { method: "POST", body: JSON.stringify(body) }),
  updateProfile: (body: { full_name?: string; avatar_url?: string; username?: string; gender?: string; birthdate?: string; country?: string }) =>
    request<{ user: User }>("/api/users/me", { method: "PATCH", body: JSON.stringify(body) }),
  changePassword: (body: { current_password: string; new_password: string }) =>
    request<{ message: string }>("/api/users/me/password", { method: "PATCH", body: JSON.stringify(body) }),
  trackDailyUse: (body: { device_id: string; apps_name: string; android_version?: string; device_brand?: string; device_model?: string; country?: string; country_code?: string; email?: string }) =>
    request<{ status: string }>("/api/tracking/daily-use", { method: "POST", body: JSON.stringify(body) }),
  getDailyStats: (params: { date: string; apps_name?: string; country_code?: string }) => {
    const q = new URLSearchParams({ date: params.date });
    if (params.apps_name) q.set("apps_name", params.apps_name);
    if (params.country_code) q.set("country_code", params.country_code);
    return request<DailyStatsResponse>(`/api/tracking/stats/daily?${q.toString()}`);
  },
  getFilterOptions: (date: string) =>
    request<FilterOptionsResponse>(`/api/tracking/filter-options?date=${encodeURIComponent(date)}`),
};

export type FilterOptionCountry = {
  country: string;
  country_code: string;
};

export type FilterOptionsResponse = {
  date: string;
  timezone: string;
  apps: string[];
  countries: FilterOptionCountry[];
};

export type HourlyBucket = {
  hour: number;
  hits: number;
  unique_devices: number;
};

export type CountryStat = {
  country: string;
  country_code: string;
  hits: number;
  unique_devices: number;
};

export type AppStat = {
  apps_name: string;
  hits: number;
  unique_devices: number;
};

export type DailyStatsResponse = {
  date: string;
  timezone: string;
  apps_name: string;
  country_code: string;
  total_hits: number;
  dau: number;
  hourly: HourlyBucket[];
  by_country: CountryStat[];
  by_app: AppStat[];
  all_apps: string[];
  all_countries: CountryStat[];
};
