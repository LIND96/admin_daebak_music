"use client";

import { Link } from "@/i18n/navigation";
import { MusicIcon } from "@/icons";
import ThemeToggleButton from "@/layout/header/theme-toggler";
import { useAuth } from "@/lib/auth-context";
import { useState } from "react";
import { navLinks, siteName } from "./data";

export default function LandingHeader() {
  const [menuOpen, setMenuOpen] = useState(false);
  const { user, loading } = useAuth();

  return (
    <header className="sticky top-0 z-50 border-b border-gray-200 bg-white/80 backdrop-blur-md dark:border-gray-800 dark:bg-gray-950/80">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <nav className="flex h-16 items-center justify-between">
          <a href="#top" className="flex items-center gap-2">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-brand-700">
              <MusicIcon className="size-5 text-white" />
            </span>
            <span className="font-outfit text-xl font-bold text-gray-900 dark:text-white">
              {siteName}
            </span>
          </a>

          <div className="hidden items-center gap-1 lg:flex">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                className="rounded-lg px-4 py-2 text-sm font-medium text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white"
              >
                {link.label}
              </a>
            ))}
          </div>

          <div className="flex items-center gap-2 sm:gap-3">
            <span className="hidden items-center gap-2 rounded-full bg-gray-100 px-3 py-1.5 sm:flex dark:bg-gray-800">
              <span className="relative flex h-2 w-2">
                <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
              </span>
              <span className="text-xs font-medium text-gray-600 dark:text-gray-400">
                All Systems Live
              </span>
            </span>

            <ThemeToggleButton />

            {!loading &&
              (user ? (
                <Link
                  href="/dashboard"
                  className="hidden rounded-lg bg-brand-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-brand-500/20 transition-colors hover:bg-brand-600 sm:block"
                >
                  Open Dashboard
                </Link>
              ) : (
                <>
                  <Link
                    href="/signin"
                    className="hidden rounded-lg px-4 py-2 text-sm font-medium text-gray-600 transition-colors hover:text-gray-900 sm:block dark:text-gray-300 dark:hover:text-white"
                  >
                    Sign in
                  </Link>
                  <Link
                    href="/signup"
                    className="hidden rounded-lg bg-brand-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-brand-500/20 transition-colors hover:bg-brand-600 sm:block"
                  >
                    Get Started
                  </Link>
                </>
              ))}

            <button
              type="button"
              onClick={() => setMenuOpen((open) => !open)}
              aria-label={menuOpen ? "Close menu" : "Open menu"}
              aria-expanded={menuOpen}
              className="rounded-lg p-2 text-gray-600 hover:bg-gray-100 lg:hidden dark:text-gray-400 dark:hover:bg-gray-800"
            >
              {menuOpen ? (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              ) : (
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
                </svg>
              )}
            </button>
          </div>
        </nav>
      </div>

      {menuOpen && (
        <nav className="border-t border-gray-200 bg-white px-4 py-4 lg:hidden dark:border-gray-800 dark:bg-gray-950">
          <div className="space-y-1">
            {navLinks.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setMenuOpen(false)}
                className="block rounded-lg px-4 py-3 font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 dark:text-gray-300 dark:hover:bg-gray-800 dark:hover:text-white"
              >
                {link.label}
              </a>
            ))}
          </div>
          <div className="mt-4 space-y-3">
            {!loading &&
              (user ? (
                <Link
                  href="/dashboard"
                  className="block rounded-lg bg-brand-500 px-4 py-3 text-center font-semibold text-white"
                >
                  Open Dashboard
                </Link>
              ) : (
                <>
                  <Link
                    href="/signin"
                    className="block rounded-lg border border-gray-200 px-4 py-3 text-center font-medium text-gray-600 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800"
                  >
                    Sign in
                  </Link>
                  <Link
                    href="/signup"
                    className="block rounded-lg bg-brand-500 px-4 py-3 text-center font-semibold text-white"
                  >
                    Get Started Free
                  </Link>
                </>
              ))}
          </div>
        </nav>
      )}
    </header>
  );
}
