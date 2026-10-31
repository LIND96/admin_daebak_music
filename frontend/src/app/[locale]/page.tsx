import LandingPage from "@/components/landing";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Daebak Music - Modern Music Admin Platform",
  description:
    "Publish tracks, monitor streams, and grow your audience from one dashboard.",
};

export default function Home() {
  return <LandingPage />;
}
