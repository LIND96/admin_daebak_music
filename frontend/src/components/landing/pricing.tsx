import { Link } from "@/i18n/navigation";
import { CheckIcon } from "@/icons";
import { cn } from "@/utils";
import { pricingTiers } from "./data";

export default function LandingPricing() {
  return (
    <section id="pricing" className="bg-white py-24 dark:bg-gray-900">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="mx-auto max-w-3xl text-center">
          <span className="mb-4 inline-block rounded-full bg-brand-50 px-3 py-1 text-xs font-semibold text-brand-600 uppercase dark:bg-brand-500/15 dark:text-brand-400">
            Pricing
          </span>
          <h2 className="font-outfit text-3xl font-bold text-gray-900 sm:text-4xl lg:text-5xl dark:text-white">
            Simple Plans That{" "}
            <span className="bg-gradient-to-r from-brand-500 to-brand-700 bg-clip-text text-transparent">
              Scale With You
            </span>
          </h2>
          <p className="mt-4 text-lg text-gray-600 dark:text-gray-400">
            Start free, upgrade when your catalog takes off. No hidden fees.
          </p>
        </div>

        <div className="mt-14 grid gap-6 lg:grid-cols-3">
          {pricingTiers.map((tier) => (
            <div
              key={tier.name}
              className={cn(
                "flex flex-col rounded-2xl border p-8",
                tier.highlighted
                  ? "border-brand-500 bg-brand-950 text-white shadow-xl shadow-brand-500/20 dark:border-brand-500"
                  : "border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950",
              )}
            >
              <h3
                className={cn(
                  "font-outfit text-lg font-semibold",
                  tier.highlighted ? "text-white" : "text-gray-900 dark:text-white",
                )}
              >
                {tier.name}
              </h3>
              <p
                className={cn(
                  "mt-1 text-sm",
                  tier.highlighted ? "text-white/70" : "text-gray-600 dark:text-gray-400",
                )}
              >
                {tier.description}
              </p>
              <p className="mt-6">
                <span
                  className={cn(
                    "font-outfit text-5xl font-extrabold",
                    tier.highlighted ? "text-white" : "text-gray-900 dark:text-white",
                  )}
                >
                  {tier.price}
                </span>
                <span
                  className={cn(
                    "ms-2 text-sm",
                    tier.highlighted ? "text-white/70" : "text-gray-500 dark:text-gray-400",
                  )}
                >
                  {tier.period}
                </span>
              </p>
              <ul className="mt-8 flex-1 space-y-3">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-3 text-sm">
                    <CheckIcon
                      className={cn(
                        "mt-0.5 size-5 shrink-0",
                        tier.highlighted
                          ? "text-green-400"
                          : "text-green-600 dark:text-green-400",
                      )}
                    />
                    <span
                      className={
                        tier.highlighted
                          ? "text-white/90"
                          : "text-gray-700 dark:text-gray-300"
                      }
                    >
                      {feature}
                    </span>
                  </li>
                ))}
              </ul>
              <Link
                href="/signup"
                className={cn(
                  "mt-8 rounded-xl px-6 py-3 text-center font-semibold transition-colors",
                  tier.highlighted
                    ? "bg-white text-brand-900 hover:bg-brand-50"
                    : "bg-brand-500 text-white hover:bg-brand-600",
                )}
              >
                {tier.cta}
              </Link>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
