"use client";

import ComponentCard from "@/components/common/component-card";
import PageBreadcrumb from "@/components/common/page-breadcrumb";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";
import { useEffect, useState } from "react";

function getInitials(name: string) {
  const initials = name
    .split(" ")
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase())
    .join("")
    .slice(0, 2);
  return initials || "?";
}

export default function ProfilePage() {
  const { user, loading, logout } = useAuth();
  const [fullName, setFullName] = useState("");
  const [username, setUsername] = useState("");
  const [avatarUrl, setAvatarUrl] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [isChangingPw, setIsChangingPw] = useState(false);
  const [pwMessage, setPwMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  useEffect(() => {
    if (loading || !user) return;
    let cancelled = false;
    api
      .me()
      .then((res) => {
        if (cancelled) return;
        setFullName(res.user.full_name);
        setUsername(res.user.username);
        setAvatarUrl(res.user.avatar_url);
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        setMessage({
          type: "error",
          text: err instanceof Error ? err.message : "Failed to load profile",
        });
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [user, loading]);

  const showLoading = loading || (!!user && isLoading);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setMessage(null);
    try {
      const res = await api.updateProfile({
        full_name: fullName,
        username,
        avatar_url: avatarUrl,
      });
      setFullName(res.user.full_name);
      setUsername(res.user.username);
      setAvatarUrl(res.user.avatar_url);
      setMessage({ type: "success", text: "Profile updated" });
    } catch (err) {
      setMessage({
        type: "error",
        text: err instanceof Error ? err.message : "Update failed",
      });
    } finally {
      setIsSaving(false);
    }
  };

  const displayName = user?.full_name || user?.username || "";

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setPwMessage(null);
    if (newPassword.length < 6) {
      setPwMessage({ type: "error", text: "Password baru minimal 6 karakter." });
      return;
    }
    if (newPassword !== confirmPassword) {
      setPwMessage({ type: "error", text: "Konfirmasi password tidak cocok." });
      return;
    }
    if (newPassword === currentPassword) {
      setPwMessage({ type: "error", text: "Password baru harus berbeda." });
      return;
    }
    setIsChangingPw(true);
    try {
      await api.changePassword({ current_password: currentPassword, new_password: newPassword });
      // Refresh token sudah di-revoke backend: sesi ini tidak bisa diperpanjang,
      // paksa login ulang dengan password baru.
      await logout();
    } catch (err) {
      setPwMessage({
        type: "error",
        text: err instanceof Error ? err.message : "Ganti password gagal",
      });
      setIsChangingPw(false);
    }
  };

  return (
    <div>
      <PageBreadcrumb pageTitle="Profile" />
      <div className="space-y-6">
        <ComponentCard title="My Profile">
          {showLoading ? (
            <p className="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              Loading profile...
            </p>
          ) : !user ? (
            <p className="py-8 text-center text-sm text-gray-500 dark:text-gray-400">
              You are not logged in.
            </p>
          ) : (
            <div className="space-y-6">
              <div className="flex items-center gap-4">
                <Avatar className="size-16">
                  <AvatarImage
                    src={avatarUrl || user?.avatar_url}
                    alt={username || "User profile"}
                  />
                  <AvatarFallback>{getInitials(displayName)}</AvatarFallback>
                </Avatar>
                <div>
                  <div className="text-lg font-medium text-gray-800 dark:text-white/90">
                    {displayName}
                  </div>
                  <div className="text-sm text-gray-500 dark:text-gray-400">
                    {user?.email}
                  </div>
                </div>
              </div>

              {message && (
                <p
                  role={message.type === "error" ? "alert" : "status"}
                  className={
                    message.type === "error"
                      ? "rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400"
                      : "rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700 dark:border-green-500/30 dark:bg-green-500/10 dark:text-green-400"
                  }
                >
                  {message.text}
                </p>
              )}

              <form onSubmit={handleSave} className="space-y-5">
                <Field>
                  <FieldLabel className="mb-1.5 block">Full Name</FieldLabel>
                  <Input
                    value={fullName}
                    onChange={(e) => setFullName(e.target.value)}
                    placeholder="Enter your full name"
                  />
                  <FieldError errors={[]} />
                </Field>

                <Field>
                  <FieldLabel className="mb-1.5 block">Username</FieldLabel>
                  <Input
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    placeholder="Enter your username"
                  />
                  <FieldError errors={[]} />
                </Field>

                <Field>
                  <FieldLabel className="mb-1.5 block">Avatar URL</FieldLabel>
                  <Input
                    value={avatarUrl}
                    onChange={(e) => setAvatarUrl(e.target.value)}
                    placeholder="https://..."
                  />
                  <FieldError errors={[]} />
                </Field>

                <Button type="submit" disabled={isSaving}>
                  {isSaving ? "Saving..." : "Save Changes"}
                </Button>
              </form>
            </div>
          )}
        </ComponentCard>

        {user && !showLoading && (
          <ComponentCard title="Change Password">
            <p className="mb-4 text-sm text-gray-500 dark:text-gray-400">
              Password lama wajib benar. Setelah berhasil, semua sesi ikut logout
              dan kamu harus login ulang dengan password baru.
            </p>
            {pwMessage && (
              <p
                role={pwMessage.type === "error" ? "alert" : "status"}
                className={
                  pwMessage.type === "error"
                    ? "mb-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-400"
                    : "mb-4 rounded-lg border border-green-200 bg-green-50 px-3 py-2 text-sm text-green-700 dark:border-green-500/30 dark:bg-green-500/10 dark:text-green-400"
                }
              >
                {pwMessage.text}
              </p>
            )}
            <form onSubmit={handleChangePassword} className="space-y-5">
              <Field>
                <FieldLabel className="mb-1.5 block">Current Password</FieldLabel>
                <Input
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="Enter current password"
                  autoComplete="current-password"
                />
                <FieldError errors={[]} />
              </Field>

              <Field>
                <FieldLabel className="mb-1.5 block">New Password (min 6)</FieldLabel>
                <Input
                  type="password"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Enter new password"
                  autoComplete="new-password"
                />
                <FieldError errors={[]} />
              </Field>

              <Field>
                <FieldLabel className="mb-1.5 block">Confirm New Password</FieldLabel>
                <Input
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Repeat new password"
                  autoComplete="new-password"
                />
                <FieldError errors={[]} />
              </Field>

              <Button type="submit" disabled={isChangingPw}>
                {isChangingPw ? "Updating..." : "Change Password"}
              </Button>
            </form>
          </ComponentCard>
        )}
      </div>
    </div>
  );
}
