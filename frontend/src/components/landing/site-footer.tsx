import { Link } from "@/i18n/navigation";
import { ArrowRightIcon, MusicIcon } from "@/icons";
import { footerColumns, siteName } from "./data";

export default function LandingFooter() {
  return (
    <footer className="bg-white dark:bg-gray-900">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="relative overflow-hidden rounded-3xl bg-brand-950 px-6 py-16 text-center sm:px-12">
          <div className="pointer-events-none absolute -top-20 start-1/2 h-64 w-96 -translate-x-1/2 rounded-full bg-brand-500/25 blur-[100px] rtl:translate-x-1/2" />
          <h2 className="relative font-outfit text-3xl font-bold text-white sm:text-4xl">
            Ready to Amplify Your Catalog?
          </h2>
          <p className="relative mx-auto mt-4 max-w-xl text-brand-100/80">
            Join thousands of artists and labels already running on Daebak
            Music. Set up in minutes, free forever to start.
          </p>
          <Link
            href="/signup"
            className="relative mt-8 inline-flex items-center gap-2 rounded-xl bg-white px-8 py-4 text-lg font-semibold text-brand-900 transition-colors hover:bg-brand-50"
          >
            Get Started Free
            <ArrowRightIcon className="size-5 rtl:rotate-180" />
          </Link>
        </div>

        <div className="grid gap-10 py-16 md:grid-cols-[1.5fr_1fr_1fr_1fr]">
          <div>
            <span className="flex items-center gap-2">
              <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-brand-700">
                <MusicIcon className="size-5 text-white" />
              </span>
              <span className="font-outfit text-xl font-bold text-gray-900 dark:text-white">
                {siteName}
              </span>
            </span>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-gray-600 dark:text-gray-400">
              The modern admin platform for music catalogs, artists, and
              listeners.
            </p>
          </div>
          {footerColumns.map((column) => (
            <div key={column.title}>
              <h3 className="font-outfit text-sm font-semibold text-gray-900 uppercase dark:text-white">
                {column.title}
              </h3>
              <ul className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <li key={link}>
                    <a
                      href="#top"
                      className="text-sm text-gray-600 transition-colors hover:text-gray-900 dark:text-gray-400 dark:hover:text-white"
                    >
                      {link}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="flex flex-col items-center justify-between gap-3 border-t border-gray-200 py-6 text-sm text-gray-500 sm:flex-row dark:border-gray-800 dark:text-gray-400">
          <p>© {new Date().getFullYear()} {siteName}. All rights reserved.</p>
          <p>Built for artists, labels, and listeners.</p>
        </div>
      </div>
    </footer>
  );
}
