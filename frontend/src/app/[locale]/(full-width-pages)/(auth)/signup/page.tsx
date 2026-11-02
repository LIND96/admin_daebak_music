import { redirect } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "SignUp Disabled | Admin Daebak Music",
  description: "Public registration is disabled. Please contact an admin.",
};

export default function SignUp() {
  // Register publik dimatikan. Admin dibuat via bootstrap / POST /api/users.
  redirect({ href: "/signin", locale: routing.defaultLocale });
}
