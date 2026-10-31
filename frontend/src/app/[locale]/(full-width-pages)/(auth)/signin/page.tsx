import SignIn from "@/components/others/auth/sign-in";
import { redirect } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import { Metadata } from "next";
import { cookies } from "next/headers";

export const metadata: Metadata = {
  title: "Next.js SignIn Page | TailAdmin - Next.js Dashboard Template",
  description: "This is Next.js Signin Page TailAdmin Dashboard Template",
};

export default async function SignInPage() {
  // Sudah login (cookie httpOnly ada) -> langsung ke dashboard.
  if ((await cookies()).has("token"))
    redirect({ href: "/dashboard", locale: routing.defaultLocale });
  return <SignIn />;
}
