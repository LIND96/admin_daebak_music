"use client";

import { useTheme } from "@/context/theme-context";
import type { HourlyBucket } from "@/lib/api";
import { ApexOptions } from "apexcharts";
import dynamic from "next/dynamic";
import { useMemo } from "react";

const ReactApexChart = dynamic(() => import("react-apexcharts"), {
  ssr: false,
});

function hourLabel(h: number) {
  return `${String(h).padStart(2, "0")}:00`;
}

export default function HourlyChart({ hourly }: { hourly: HourlyBucket[] }) {
  const { theme } = useTheme();

  const categories = useMemo(
    () => Array.from({ length: 24 }, (_, h) => hourLabel(h)),
    [],
  );

  const options: ApexOptions = useMemo(
    () => ({
      legend: { show: false },
      colors: ["#465FFF"],
      chart: {
        fontFamily: "Outfit, sans-serif",
        height: 310,
        type: "area",
        toolbar: { show: false },
      },
      stroke: { curve: "smooth", width: 2 },
      fill: {
        type: "gradient",
        gradient: { opacityFrom: 0.55, opacityTo: 0 },
      },
      markers: {
        size: 0,
        colors: ["#ffffff"],
        strokeColors: ["#465FFF"],
        strokeWidth: 2,
        hover: { size: 5 },
      },
      grid: {
        position: "back",
        borderColor: theme === "dark" ? "#1E2640" : "#F3F4F6",
        strokeDashArray: 0,
        xaxis: { lines: { show: false } },
        yaxis: { lines: { show: true } },
      },
      dataLabels: { enabled: false },
      tooltip: { enabled: true, shared: true, intersect: false },
      xaxis: {
        type: "category",
        categories,
        axisBorder: { show: false },
        axisTicks: { show: false },
        crosshairs: {
          show: true,
          width: 2,
          position: "back",
          opacity: 1,
          stroke: { width: 0, color: "none" },
          fill: {
            type: "gradient",
            gradient: {
              colorFrom: theme === "dark" ? "rgba(255, 255, 255, 0.1)" : "#ffffff",
              colorTo: "#9CB9FF",
              stops: [0, 100],
              opacityFrom: 1,
              opacityTo: 1,
            },
          },
        },
        tooltip: { enabled: false },
        labels: {
          style: {
            fontSize: "12px",
            fontFamily: "Outfit, sans-serif",
            colors: theme === "dark" ? "#94A3B8" : "#6B7280",
          },
        },
      },
      yaxis: {
        min: 0,
        labels: {
          formatter: (val: number): string => {
            const n = Math.round(val);
            if (n >= 1000) return `${(n / 1000).toFixed(n % 1000 === 0 ? 0 : 1)}K`;
            return `${n}`;
          },
          style: {
            fontSize: "12px",
            fontFamily: "Outfit, sans-serif",
            colors: theme === "dark" ? "#94A3B8" : "#6B7280",
          },
        },
        title: { text: undefined },
      },
    }),
    [categories, theme],
  );

  const series = useMemo(
    () => [
      {
        name: "Total hits",
        data: Array.from({ length: 24 }, (_, h) => hourly[h]?.hits ?? 0),
      },
    ],
    [hourly],
  );

  return (
    <div className="rounded-2xl border border-gray-200 bg-white px-5 pt-5 pb-5 sm:px-6 sm:pt-6 dark:border-gray-800 dark:bg-white/3">
      <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">
        Total hits per jam (WIB)
      </h3>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Semua tembakan tercatat, 00:00–23:00
      </p>
      <div className="mt-4 max-w-full overflow-x-auto">
        <div className="min-w-150 xl:min-w-full">
          <ReactApexChart options={options} series={series} type="area" height={310} />
        </div>
      </div>
    </div>
  );
}
