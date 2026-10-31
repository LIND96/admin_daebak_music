import LandingFaq from "./faq";
import LandingFeatures from "./features";
import LandingHero from "./hero";
import PhoneMockup from "./phone-mockup";
import LandingPricing from "./pricing";
import LandingShowcase from "./showcase";
import LandingFooter from "./site-footer";
import LandingHeader from "./site-header";
import StatsBar from "./stats-bar";
import LandingTicker from "./ticker";

export default function LandingPage() {
  return (
    <div id="top" className="min-h-screen bg-gray-50 font-outfit dark:bg-gray-950">
      <LandingTicker />
      <LandingHeader />
      <main>
        <section className="relative overflow-hidden pb-16 pt-24 lg:pb-24 lg:pt-32">
          <div className="absolute inset-0 bg-[radial-gradient(circle_at_1px_1px,rgba(100,116,139,0.18)_1px,transparent_0)] bg-[size:24px_24px] dark:bg-[radial-gradient(circle_at_1px_1px,rgba(148,163,184,0.12)_1px,transparent_0)]" />
          <div className="pointer-events-none absolute end-0 top-0 h-[600px] w-[600px] rounded-full bg-brand-500/10 blur-[120px]" />
          <div className="pointer-events-none absolute bottom-0 start-0 h-[400px] w-[400px] rounded-full bg-brand-500/10 blur-[100px]" />
          <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid items-center gap-12 lg:grid-cols-2 lg:gap-16">
              <LandingHero />
              <PhoneMockup />
            </div>
            <StatsBar />
          </div>
        </section>

        <LandingFeatures />
        <LandingShowcase />
        <LandingPricing />
        <LandingFaq />
      </main>
      <LandingFooter />
    </div>
  );
}
