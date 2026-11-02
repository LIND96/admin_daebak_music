"use client";

import { useTheme } from "@/context/theme-context";
import type { AppStat } from "@/lib/api";
import { ApexOptions } from "apexcharts";
import dynamic from "next/dynamic";
import { useMemo } from "react";

const ReactApexChart = dynamic(() => import("react-apexcharts"), {
  ssr: false,
});

const APP_COLORS = [
  "#465FFF",
  "#12B76A",
  "#F79009",
  "#EE46BC",
  "#9E77ED",
  "#06BEE1",
  "#F04438",
  "#667085",
  "#84E1B6",
  "#BAD2FF",
];

export default function AppsChart({ data }: { data: AppStat[] }) {
  const { theme } = useTheme();

  const categories = useMemo(() => data.map((d) => d.apps_name), [data]);

  const barColors = useMemo(
    () => data.map((_, i) => APP_COLORS[i % APP_COLORS.length]),
    [data],
  );

  const options: ApexOptions = useMemo(
    () => ({
      colors: barColors.length > 0 ? barColors : APP_COLORS,
      chart: {
        fontFamily: "Outfit, sans-serif",
        type: "bar",
        height: Math.max(220, data.length * 52 + 60),
        toolbar: { show: false },
      },
      plotOptions: {
        bar: {
          horizontal: true,
          distributed: true,
          borderRadius: 6,
          borderRadiusApplication: "end",
        },
      },
      dataLabels: { enabled: false },
      xaxis: {
        categories,
        labels: {
          style: {
            colors: theme === "dark" ? "#94A3B8" : "#6B7280",
            fontFamily: "Outfit, sans-serif",
          },
        },
      },
      yaxis: {
        labels: {
          style: {
            colors: theme === "dark" ? "#94A3B8" : "#364153",
            fontFamily: "Outfit, sans-serif",
            fontSize: "12px",
          },
        },
      },
      grid: {
        borderColor: theme === "dark" ? "#1E2640" : "#F3F4F6",
        xaxis: { lines: { show: true } },
        yaxis: { lines: { show: false } },
      },
      legend: { show: false },
      tooltip: { shared: true, intersect: false },
    }),
    [categories, data.length, barColors, theme],
  );

  const series = useMemo(
    () => [{ name: "User aktif (unik device)", data: data.map((d) => d.unique_devices) }],
    [data],
  );

  if (data.length === 0) {
    return (
      <div className="rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
        <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">
          User aktif per aplikasi
        </h3>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          Belum ada data untuk tanggal ini.
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-gray-200 bg-white px-5 pt-5 pb-5 sm:px-6 sm:pt-6 dark:border-gray-800 dark:bg-white/3">
      <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">
        User aktif per aplikasi
      </h3>
      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
        Unik berdasarkan device_id per apps_name · {data.length} aplikasi
      </p>
      <div className="mt-4">
        <ReactApexChart
          options={options}
          series={series}
          type="bar"
          height={Math.max(220, data.length * 52 + 60)}
        />
      </div>
      <ul className="mt-4 space-y-2">
        {data.map((d, i) => (
          <li
            key={d.apps_name}
            className="flex items-center justify-between text-sm text-gray-600 dark:text-gray-300"
          >
            <span className="flex items-center gap-2 font-medium">
              <span
                className="inline-block size-3 rounded-full"
                style={{ backgroundColor: APP_COLORS[i % APP_COLORS.length] }}
              />
              {d.apps_name}
            </span>
            <span>{d.unique_devices.toLocaleString("id-ID")} user unik</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
