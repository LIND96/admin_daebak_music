import SignUpForm from "@/components/others/auth/sign-up";
import { redirect } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import { Metadata } from "next";
import { cookies } from "next/headers";

export const metadata: Metadata = {
  title: "Next.js SignUp Page | TailAdmin - Next.js Dashboard Template",
  description: "This is Next.js SignUp Page TailAdmin Dashboard Template",
  // other metadata
};

export default async function SignUp() {
  // Sudah login (cookie httpOnly ada) -> langsung ke dashboard.
  if ((await cookies()).has("token"))
    redirect({ href: "/dashboard", locale: routing.defaultLocale });
  return <SignUpForm />;
}
