import { BellIcon, MusicIcon, PlayIcon, SearchIcon, UsersIcon } from "@/icons";

const topTracks = [
  { initial: "M", title: "Midnight Drive", meta: "128K plays", plays: "1.2M", change: "+4.2%", up: true, chip: "bg-brand-100 text-brand-600 dark:bg-brand-500/15 dark:text-brand-400" },
  { initial: "S", title: "Solar Dreams", meta: "96K plays", plays: "864K", change: "+7.9%", up: true, chip: "bg-orange-100 text-orange-600 dark:bg-orange-500/15 dark:text-orange-400" },
  { initial: "N", title: "Neon Hearts", meta: "74K plays", plays: "702K", change: "-1.3%", up: false, chip: "bg-green-100 text-green-600 dark:bg-green-500/15 dark:text-green-400" },
  { initial: "O", title: "Ocean Eyes", meta: "61K plays", plays: "588K", change: "+2.4%", up: true, chip: "bg-purple-100 text-purple-600 dark:bg-purple-500/15 dark:text-purple-400" },
];

export default function PhoneMockup() {
  return (
    <div className="relative">
      <div className="relative mx-auto max-w-md lg:max-w-none">
        <div className="relative mx-auto aspect-[9/19.5] w-64 rounded-[3rem] border-8 border-gray-900 bg-gray-900 p-2 shadow-2xl lg:w-72 dark:border-gray-700">
          <div className="absolute start-1/2 top-4 z-10 h-7 w-24 -translate-x-1/2 rounded-full bg-gray-900 rtl:translate-x-1/2 dark:bg-gray-700" />
          <div className="flex h-full flex-col overflow-hidden rounded-[2.25rem] bg-white dark:bg-gray-900">
            <div className="bg-gradient-to-r from-brand-500 to-brand-700 px-4 pb-4 pt-10 text-white">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-xs opacity-80">Streams Today</p>
                  <p className="font-mono text-xl font-bold">1,284,392</p>
                  <p className="flex items-center gap-1 text-xs text-green-200">
                    <PlayIcon className="size-3" />
                    +12,480 (+5.2%) today
                  </p>
                </div>
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-white/20">
                  <BellIcon className="size-4" />
                </span>
              </div>
            </div>

            <div className="relative h-28 bg-gray-50 dark:bg-gray-800">
              <div className="absolute start-0 end-0 top-2 flex justify-center gap-2 px-4">
                {["1D", "1W", "1M", "1Y", "ALL"].map((tab) => (
                  <span
                    key={tab}
                    className={
                      tab === "1D"
                        ? "rounded-full bg-brand-500 px-2 py-0.5 text-[10px] font-medium text-white"
                        : "rounded-full px-2 py-0.5 text-[10px] font-medium text-gray-500 dark:text-gray-400"
                    }
                  >
                    {tab}
                  </span>
                ))}
              </div>
              <svg className="h-full w-full pt-6" viewBox="0 0 300 100" preserveAspectRatio="none" aria-hidden>
                <defs>
                  <linearGradient id="landingSpark" x1="0%" y1="0%" x2="0%" y2="100%">
                    <stop offset="0%" stopColor="#465fff" stopOpacity="0.3" />
                    <stop offset="100%" stopColor="#465fff" stopOpacity="0" />
                  </linearGradient>
                </defs>
                <path d="M0,82 Q30,74 60,66 T120,52 T180,58 T240,32 T300,22 V100 H0 Z" fill="url(#landingSpark)" />
                <path d="M0,82 Q30,74 60,66 T120,52 T180,58 T240,32 T300,22" fill="none" stroke="#465fff" strokeWidth="2" />
              </svg>
            </div>

            <div className="flex-1 overflow-hidden">
              <div className="flex items-center justify-between border-b border-gray-100 px-4 py-2 dark:border-gray-800">
                <span className="text-xs font-semibold text-gray-900 dark:text-white">Top Tracks</span>
                <span className="text-xs text-brand-600 dark:text-brand-400">See all</span>
              </div>
              <div className="divide-y divide-gray-100 px-4 dark:divide-gray-800">
                {topTracks.map((track) => (
                  <div key={track.title} className="flex items-center justify-between py-2.5">
                    <div className="flex items-center gap-2.5">
                      <span className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${track.chip}`}>
                        {track.initial}
                      </span>
                      <div>
                        <p className="text-sm font-semibold text-gray-900 dark:text-white">{track.title}</p>
                        <p className="text-[10px] text-gray-500">{track.meta}</p>
                      </div>
                    </div>
                    <div className="text-end">
                      <p className="font-mono text-sm font-semibold text-gray-900 dark:text-white">{track.plays}</p>
                      <p className={track.up ? "text-[10px] text-green-600 dark:text-green-400" : "text-[10px] text-red-600 dark:text-red-400"}>
                        {track.change}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="border-t border-gray-100 bg-white px-6 py-3 dark:border-gray-800 dark:bg-gray-900">
              <div className="flex items-center justify-between">
                <span className="flex flex-col items-center gap-1">
                  <MusicIcon className="size-5 text-brand-600 dark:text-brand-400" />
                  <span className="text-[9px] font-medium text-brand-600 dark:text-brand-400">Home</span>
                </span>
                <span className="flex flex-col items-center gap-1">
                  <SearchIcon className="size-5 text-gray-400" />
                  <span className="text-[9px] font-medium text-gray-400">Search</span>
                </span>
                <span className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-500">
                  <PlayIcon className="size-4 text-white" />
                </span>
                <span className="flex flex-col items-center gap-1">
                  <BellIcon className="size-5 text-gray-400" />
                  <span className="text-[9px] font-medium text-gray-400">Alerts</span>
                </span>
                <span className="flex flex-col items-center gap-1">
                  <UsersIcon className="size-5 text-gray-400" />
                  <span className="text-[9px] font-medium text-gray-400">Fans</span>
                </span>
              </div>
            </div>
            <div className="flex justify-center bg-white pb-2 dark:bg-gray-900">
              <div className="h-1 w-24 rounded-full bg-gray-300 dark:bg-gray-600" />
            </div>
          </div>
        </div>

        <div className="absolute -start-4 top-24 hidden animate-float rounded-xl border border-gray-200 bg-white p-3 shadow-lg lg:block dark:border-gray-700 dark:bg-gray-800" style={{ animationDelay: "0.5s" }}>
          <div className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-full bg-green-100 dark:bg-green-500/15">
              <MusicIcon className="size-4 text-green-600 dark:text-green-400" />
            </span>
            <div>
              <p className="text-[10px] text-gray-500">Track Published</p>
              <p className="text-sm font-semibold text-gray-900 dark:text-white">Midnight Drive EP</p>
            </div>
          </div>
        </div>
        <div className="absolute -end-2 bottom-44 hidden animate-float rounded-xl border border-gray-200 bg-white p-3 shadow-lg lg:block dark:border-gray-700 dark:bg-gray-800" style={{ animationDelay: "1s" }}>
          <div className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-full bg-brand-100 dark:bg-brand-500/15">
              <PlayIcon className="size-4 text-brand-600 dark:text-brand-400" />
            </span>
            <div>
              <p className="text-[10px] text-gray-500">Milestone Reached</p>
              <p className="text-sm font-semibold text-gray-900 dark:text-white">1M streams today</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
