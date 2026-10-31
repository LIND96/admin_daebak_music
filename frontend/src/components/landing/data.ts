export const siteName = "Daebak Music";

export const navLinks = [
  { label: "Features", href: "#features" },
  { label: "Showcase", href: "#showcase" },
  { label: "Pricing", href: "#pricing" },
  { label: "FAQ", href: "#faq" },
];

export type TickerItem = {
  symbol: string;
  value: string;
  change: string;
  up: boolean;
};

export const tickerItems: TickerItem[] = [
  { symbol: "TOP 50", value: "128.4M streams", change: "+4.2%", up: true },
  { symbol: "NEW RELEASES", value: "1,842 tracks", change: "+11.7%", up: true },
  { symbol: "ACTIVE FANS", value: "2.1M", change: "+6.3%", up: true },
  { symbol: "VINYL", value: "48.2K orders", change: "-1.8%", up: false },
  { symbol: "PLAYLISTS", value: "96.7K", change: "+9.4%", up: true },
  { symbol: "LIVE TIX", value: "12.3K sold", change: "+3.1%", up: true },
  { symbol: "PODCASTS", value: "312 shows", change: "+7.9%", up: true },
  { symbol: "ROYALTIES", value: "$842K paid", change: "-0.6%", up: false },
];

export type Stat = { value: string; label: string };

export const heroStats: Stat[] = [
  { value: "120K+", label: "Tracks Managed" },
  { value: "8.4M", label: "Monthly Listeners" },
  { value: "96", label: "Label Partners" },
  { value: "24/7", label: "Platform Uptime" },
];

export type Feature = {
  icon: "users" | "chart" | "bolt" | "bell" | "globe" | "music";
  title: string;
  description: string;
};

export const features: Feature[] = [
  {
    icon: "users",
    title: "Artist & User Management",
    description:
      "Manage artists, listeners, and staff accounts with role-based access from one clean dashboard.",
  },
  {
    icon: "chart",
    title: "Streaming Analytics",
    description:
      "Track plays, skips, and listener growth per track, playlist, and region in real time.",
  },
  {
    icon: "bolt",
    title: "Instant Publishing",
    description:
      "Release new tracks and update catalogs in seconds with zero downtime for listeners.",
  },
  {
    icon: "bell",
    title: "Smart Notifications",
    description:
      "Get alerted on milestones, trending tracks, and unusual activity as it happens.",
  },
  {
    icon: "globe",
    title: "Global Reach",
    description:
      "Serve listeners across regions with localized content and multi-language support.",
  },
  {
    icon: "music",
    title: "Catalog Control",
    description:
      "Organize thousands of tracks, albums, and playlists with powerful search and filters.",
  },
];

export const showcasePoints = [
  "Live listener counts and stream charts update every second",
  "One-click publishing with scheduled release dates",
  "Royalty-ready reports exportable to CSV in one click",
];

export type PricingTier = {
  name: string;
  price: string;
  period: string;
  description: string;
  cta: string;
  highlighted: boolean;
  features: string[];
};

export const pricingTiers: PricingTier[] = [
  {
    name: "Starter",
    price: "$0",
    period: "forever",
    description: "For indie artists getting their first tracks online.",
    cta: "Start for free",
    highlighted: false,
    features: [
      "Up to 50 tracks",
      "Basic stream analytics",
      "1 staff account",
      "Community support",
    ],
  },
  {
    name: "Pro",
    price: "$19",
    period: "per month",
    description: "For growing labels that need deeper insight and control.",
    cta: "Start 14-day trial",
    highlighted: true,
    features: [
      "Unlimited tracks",
      "Real-time analytics",
      "10 staff accounts",
      "Scheduled releases",
      "CSV royalty reports",
      "Priority support",
    ],
  },
  {
    name: "Enterprise",
    price: "Custom",
    period: "annual billing",
    description: "For large catalogs with custom compliance needs.",
    cta: "Talk to sales",
    highlighted: false,
    features: [
      "Everything in Pro",
      "Unlimited staff accounts",
      "SSO & audit logs",
      "Dedicated manager",
      "Custom SLA",
    ],
  },
];

export type Faq = { question: string; answer: string };

export const faqs: Faq[] = [
  {
    question: "What is Daebak Music?",
    answer:
      "Daebak Music is an admin platform for running a music catalog: manage artists, publish tracks, monitor streams, and understand listeners from a single dashboard.",
  },
  {
    question: "Do I need a credit card to try it?",
    answer:
      "No. The Starter plan is free forever and the Pro trial runs for 14 days without asking for payment details.",
  },
  {
    question: "Can my whole team use one account?",
    answer:
      "Yes. Invite staff with roles such as admin, editor, or viewer, so everyone sees exactly what they need and nothing more.",
  },
  {
    question: "How does authentication work?",
    answer:
      "Accounts are protected with hashed passwords and short-lived signed sessions stored in secure httpOnly cookies.",
  },
  {
    question: "Can I export my data?",
    answer:
      "Yes. Stream reports and user lists can be exported to CSV at any time, and your catalog remains yours.",
  },
  {
    question: "Is there a dark mode?",
    answer:
      "Yes. The entire interface, including this page, supports light and dark themes with one click.",
  },
];

export const footerColumns = [
  {
    title: "Product",
    links: ["Features", "Pricing", "Changelog", "Roadmap"],
  },
  {
    title: "Resources",
    links: ["Documentation", "API Reference", "Status", "Support"],
  },
  {
    title: "Company",
    links: ["About", "Blog", "Careers", "Contact"],
  },
];
