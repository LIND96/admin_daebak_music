import {
  BellIcon,
  BoltIcon,
  GlobeIcon,
  MusicIcon,
  PieChartIcon,
  UsersIcon,
} from "@/icons";
import { features, type Feature } from "./data";

const iconMap: Record<Feature["icon"], (props: React.SVGProps<SVGSVGElement>) => React.JSX.Element> = {
  users: UsersIcon,
  chart: PieChartIcon,
  bolt: BoltIcon,
  bell: BellIcon,
  globe: GlobeIcon,
  music: MusicIcon,
};

export default function LandingFeatures() {
  return (
    <section id="features" className="bg-white py-24 dark:bg-gray-900">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-3xl text-center">
          <span className="mb-4 inline-block rounded-full bg-brand-50 px-3 py-1 text-xs font-semibold text-brand-600 uppercase dark:bg-brand-500/15 dark:text-brand-400">
            Platform Features
          </span>
          <h2 className="font-outfit text-3xl font-bold text-gray-900 sm:text-4xl lg:text-5xl dark:text-white">
            Everything You Need to{" "}
            <span className="bg-gradient-to-r from-brand-500 to-brand-700 bg-clip-text text-transparent">
              Run Your Label
            </span>
          </h2>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            Powerful tools designed for artists, managers, and label teams of
            every size.
          </p>
        </div>

        <div className="mt-14 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => {
            const Icon = iconMap[feature.icon];
            return (
              <div
                key={feature.title}
                className="rounded-2xl border border-gray-200 bg-gray-50 p-6 transition-shadow hover:shadow-lg dark:border-gray-800 dark:bg-gray-950"
              >
                <span className="inline-flex h-12 w-12 items-center justify-center rounded-xl bg-brand-500/10 text-brand-600 dark:text-brand-400">
                  <Icon className="size-6" />
                </span>
                <h3 className="mt-4 font-outfit text-lg font-semibold text-gray-900 dark:text-white">
                  {feature.title}
                </h3>
                <p className="mt-2 text-sm leading-relaxed text-gray-600 dark:text-gray-400">
                  {feature.description}
                </p>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
