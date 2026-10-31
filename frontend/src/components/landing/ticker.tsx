import { tickerItems } from "./data";

function TickerRow({ ariaHidden = false }: { ariaHidden?: boolean }) {
  return (
    <div className="flex shrink-0 items-center" aria-hidden={ariaHidden}>
      {tickerItems.map((item) => (
        <div key={item.symbol} className="flex shrink-0 items-center gap-3 px-4">
          <span className="font-mono text-sm font-medium text-gray-500 dark:text-gray-400">
            {item.symbol}
          </span>
          <span className="font-mono text-sm text-gray-900 dark:text-white">
            {item.value}
          </span>
          <span
            className={
              item.up
                ? "font-mono text-sm text-green-600 dark:text-green-400"
                : "font-mono text-sm text-red-600 dark:text-red-400"
            }
          >
            {item.change}
          </span>
        </div>
      ))}
    </div>
  );
}

export default function LandingTicker() {
  return (
    <div className="border-b border-gray-200 bg-gray-100 dark:border-gray-800 dark:bg-gray-900">
      <div className="mx-auto max-w-7xl overflow-hidden px-4">
        <div className="flex w-max animate-ticker items-center gap-8 py-2 hover:[animation-play-state:paused]">
          <TickerRow />
          <TickerRow ariaHidden />
        </div>
      </div>
    </div>
  );
}
