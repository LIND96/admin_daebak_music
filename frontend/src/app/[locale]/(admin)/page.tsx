import { redirect } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";

export default function AdminHome() {
  redirect({ href: "/dashboard", locale: routing.defaultLocale });
}
