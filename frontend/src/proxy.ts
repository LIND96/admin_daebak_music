import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import createMiddleware from "next-intl/middleware";

import { routing } from "./i18n/routing";

const intlMiddleware = createMiddleware(routing);

// Halaman auth: user yang sudah login (punya cookie) dialihkan ke dashboard.
const guestOnlyRoutes = ["/signin", "/signup"];

// Halaman yang boleh dibuka tanpa login.
const publicRoutes = ["/", "/signin", "/signup", "/error-404"];

function matches(pathname: string, routes: string[]) {
  return routes.some(
    (route) =>
      pathname === route ||
      (route !== "/" && pathname.startsWith(`${route}/`)),
  );
}

export function proxy(request: NextRequest) {
  const intlResponse = intlMiddleware(request);

  // Cek optimis: hanya keberadaan cookie. Verifikasi JWT tetap di backend Go.
  const token = request.cookies.get("token")?.value;
  const { pathname } = request.nextUrl;

  if (!token && !matches(pathname, publicRoutes)) {
    const url = request.nextUrl.clone();
    url.pathname = "/signin";
    return NextResponse.redirect(url);
  }

  if (token && matches(pathname, guestOnlyRoutes)) {
    const url = request.nextUrl.clone();
    url.pathname = "/dashboard";
    return NextResponse.redirect(url);
  }

  return intlResponse;
}

export const config = {
  matcher: ["/((?!api/|_next|_vercel|.*\\..*).*)"],
};
