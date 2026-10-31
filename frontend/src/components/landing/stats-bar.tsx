import { heroStats } from "./data";

export default function StatsBar() {
  return (
    <div className="mt-16 grid grid-cols-2 gap-6 rounded-2xl border border-gray-200 bg-white p-6 sm:grid-cols-4 lg:mt-24 dark:border-gray-800 dark:bg-gray-900">
      {heroStats.map((stat) => (
        <div key={stat.label} className="text-center">
          <p className="font-mono text-3xl font-bold text-gray-900 dark:text-white">
            {stat.value}
          </p>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            {stat.label}
          </p>
        </div>
      ))}
    </div>
  );
}
