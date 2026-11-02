import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import createMiddleware from "next-intl/middleware";

import { routing } from "./i18n/routing";

const intlMiddleware = createMiddleware(routing);

// Halaman auth: user yang sudah login (punya cookie) dialihkan ke dashboard.
// /signup dimatikan (register publik off) — selalu lempar ke /signin.
const guestOnlyRoutes = ["/signin"];

// Halaman yang boleh dibuka tanpa login.
const publicRoutes = ["/", "/signin", "/error-404"];

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

  // /signup dimatikan: register publik off, admin dibuat via bootstrap / API.
  if (pathname === "/signup" || pathname.startsWith("/signup/")) {
    const url = request.nextUrl.clone();
    url.pathname = "/signin";
    return NextResponse.redirect(url);
  }

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
