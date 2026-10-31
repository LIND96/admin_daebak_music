import { Link } from "@/i18n/navigation";
import { ArrowRightIcon, CheckCircleIcon } from "@/icons";
import { showcasePoints } from "./data";

const weekBars = [38, 55, 44, 70, 62, 88, 76];

export default function LandingShowcase() {
  return (
    <section id="showcase" className="relative overflow-hidden bg-brand-950 py-24">
      <div className="pointer-events-none absolute -top-24 start-1/4 h-96 w-96 rounded-full bg-brand-500/20 blur-[120px]" />
      <div className="relative mx-auto grid max-w-7xl items-center gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:gap-16 lg:px-8">
        <div className="text-center lg:text-start">
          <span className="mb-4 inline-block rounded-full bg-white/10 px-3 py-1 text-xs font-semibold text-brand-200 uppercase">
            Live Dashboard
          </span>
          <h2 className="font-outfit text-3xl font-bold text-white sm:text-4xl lg:text-5xl">
            Watch Your Catalog Perform in Real Time
          </h2>
          <p className="mt-4 text-lg text-brand-100/80">
            Every play, playlist add, and new follower lands on your dashboard
            the second it happens.
          </p>
          <ul className="mt-8 space-y-4 text-start">
            {showcasePoints.map((point) => (
              <li key={point} className="flex items-start gap-3">
                <CheckCircleIcon className="mt-0.5 size-6 shrink-0 text-green-400" />
                <span className="text-white/90">{point}</span>
              </li>
            ))}
          </ul>
          <Link
            href="/signup"
            className="mt-8 inline-flex items-center gap-2 rounded-xl bg-white px-8 py-4 text-lg font-semibold text-brand-900 transition-colors hover:bg-brand-50"
          >
            Try It Free
            <ArrowRightIcon className="size-5 rtl:rotate-180" />
          </Link>
        </div>

        <div className="rounded-2xl border border-white/10 bg-white/5 p-6 backdrop-blur">
          <div className="flex items-center justify-between">
            <p className="font-outfit text-sm font-semibold text-white">
              Weekly Streams
            </p>
            <span className="rounded-full bg-green-400/15 px-2.5 py-1 font-mono text-xs font-medium text-green-300">
              +18.2%
            </span>
          </div>
          <p className="mt-1 font-mono text-3xl font-bold text-white">
            8,412,906
          </p>
          <div className="mt-6 flex h-40 items-end gap-2" aria-hidden>
            {weekBars.map((height, i) => (
              <div
                key={i}
                style={{ height: `${height}%` }}
                className="flex-1 rounded-t-md bg-gradient-to-t from-brand-600 to-brand-400"
              />
            ))}
          </div>
          <div className="mt-2 flex justify-between font-mono text-[10px] text-white/50">
            {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((day) => (
              <span key={day} className="flex-1 text-center">
                {day}
              </span>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
