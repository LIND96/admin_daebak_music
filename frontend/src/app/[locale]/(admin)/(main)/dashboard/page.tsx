import ChatSection from "@/components/main/dashboard/tracking/chat-section";
import TrackingSection from "@/components/main/dashboard/tracking/tracking-section";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Dashboard | Daily App Usage",
  description: "Daily active users, hits per hour, countries and apps.",
};

export default function Dashboard() {
  return (
    <div className="grid grid-cols-12 gap-4 md:gap-6">
      <TrackingSection />
      <ChatSection />
    </div>
  );
}
