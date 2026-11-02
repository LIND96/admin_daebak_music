import { Link } from "@/i18n/navigation";
import { ArrowRightIcon, PlayIcon, StarIcon } from "@/icons";

const avatars = [
  { initials: "AR", className: "bg-brand-500" },
  { initials: "JT", className: "bg-green-500" },
  { initials: "SK", className: "bg-orange-500" },
  { initials: "LM", className: "bg-purple-500" },
];

export default function LandingHero() {
  return (
    <div className="text-center lg:text-start">
      <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-brand-500/30 bg-brand-500/10 px-4 py-2">
        <span className="relative flex h-2 w-2">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-brand-400 opacity-75" />
          <span className="relative inline-flex h-2 w-2 rounded-full bg-brand-500" />
        </span>
        <span className="text-sm font-medium text-brand-700 dark:text-brand-300">
          New: Real-time Analytics v2
        </span>
      </div>

      <h1 className="font-outfit text-4xl font-extrabold tracking-tight text-gray-900 sm:text-5xl lg:text-6xl dark:text-white">
        Manage Smarter,
        <br />
        <span className="bg-gradient-to-r from-brand-500 to-brand-700 bg-clip-text text-transparent">
          Grow Louder
        </span>
      </h1>

      <p className="mx-auto mt-6 max-w-xl text-lg text-gray-600 lg:mx-0 dark:text-gray-400">
        The modern admin platform for music catalogs, artists, and listeners.
        Publish tracks, monitor streams, and grow your audience from one
        dashboard.
      </p>

      <div className="mt-8 flex flex-col items-center gap-4 sm:flex-row lg:justify-start">
        <Link
          href="/dashboard"
          className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-brand-500 px-8 py-4 text-lg font-semibold text-white shadow-lg shadow-brand-500/25 transition-colors hover:bg-brand-600 sm:w-auto"
        >
          Open Dashboard
          <ArrowRightIcon className="size-5 rtl:rotate-180" />
        </Link>
        <a
          href="#showcase"
          className="inline-flex w-full items-center justify-center gap-2 rounded-xl border border-gray-300 bg-white px-8 py-4 text-lg font-semibold text-gray-900 transition-colors hover:bg-gray-50 sm:w-auto dark:border-gray-700 dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
        >
          <PlayIcon className="size-5" />
          See How It Works
        </a>
      </div>

      <div className="mt-10 flex flex-col items-center gap-6 lg:flex-row lg:items-center">
        <div className="flex -space-x-3">
          {avatars.map((avatar) => (
            <span
              key={avatar.initials}
              className={`flex h-10 w-10 items-center justify-center rounded-full border-2 border-white text-xs font-bold text-white dark:border-gray-950 ${avatar.className}`}
            >
              {avatar.initials}
            </span>
          ))}
          <span className="flex h-10 w-10 items-center justify-center rounded-full border-2 border-white bg-gray-900 text-xs font-bold text-white dark:border-gray-950 dark:bg-gray-700">
            2K+
          </span>
        </div>
        <div className="flex items-center gap-1">
          <span className="flex text-yellow-400">
            {Array.from({ length: 5 }).map((_, i) => (
              <StarIcon key={i} className="size-5 fill-current" />
            ))}
          </span>
          <span className="ms-2 text-sm text-gray-600 dark:text-gray-400">
            4.9/5 from 2,000+ label reviews
          </span>
        </div>
      </div>
    </div>
  );
}
