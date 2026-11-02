"use client";

import { useRouter } from "@/i18n/navigation";
import React, { createContext, useContext, useEffect, useState } from "react";
import { api, User } from "./api";

type AuthContextType = {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  /** Register publik dimatikan; tetap ada agar form signup lama tidak crash, selalu throw. */
  register: (data: {
    email: string;
    username: string;
    password: string;
    full_name?: string;
  }) => Promise<void>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

/** Web admin-only: hanya role admin yang boleh bertahan di sesi web. */
function assertAdmin(user: User) {
  if (user.role !== "admin") {
    throw new Error("Akun ini bukan admin. Login admin wajib email+password akun admin.");
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    // Cookie httpOnly terkirim otomatis; tidak perlu cek localStorage.
    api
      .me()
      .then((res) => {
        try {
          assertAdmin(res.user);
          setUser(res.user);
        } catch {
          setUser(null);
        }
      })
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const login = async (email: string, password: string) => {
    const res = await api.login({ email, password });
    assertAdmin(res.user);
    setUser(res.user);
  };

  const register = async () => {
    throw new Error("Registrasi publik dimatikan. Minta admin untuk buatkan akun.");
  };

  const logout = async () => {
    try {
      await api.logout();
    } catch {
      // abaikan, tetap bersihkan state client
    } finally {
      setUser(null);
      router.push("/signin");
    }
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
