"use client";

import { api, type DailyStatsResponse } from "@/lib/api";
import { useCallback, useEffect, useMemo, useState } from "react";
import AppsChart from "./apps-chart";
import CountriesChart from "./countries-chart";
import HourlyChart from "./hourly-chart";

function todayWIB(): string {
  const fmt = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  return fmt.format(new Date());
}

const emptyHourly = Array.from({ length: 24 }, (_, h) => ({
  hour: h,
  hits: 0,
  unique_devices: 0,
}));

export default function TrackingSection() {
  const [date, setDate] = useState<string>(todayWIB());
  const [appsName, setAppsName] = useState<string>("");
  const [appliedApps, setAppliedApps] = useState<string>("");
  const [data, setData] = useState<DailyStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");

  const fetchStats = useCallback(async (d: string, app: string) => {
    setLoading(true);
    setError("");
    try {
      const res = await api.getDailyStats({ date: d, apps_name: app || undefined });
      // Ensure 24 buckets even if backend returns fewer.
      const byHour = new Map(res.hourly.map((b) => [b.hour, b]));
      const hourly = Array.from({ length: 24 }, (_, h) => byHour.get(h) ?? emptyHourly[h]);
      setData({ ...res, hourly });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats(date, appliedApps);
  }, [date, appliedApps, fetchStats]);

  const avgPerDevice = useMemo(() => {
    if (!data || data.dau === 0) return "0";
    return (data.total_hits / data.dau).toFixed(1);
  }, [data]);

  const shiftDay = (delta: number) => {
    const [y, m, d] = date.split("-").map(Number);
    const dt = new Date(Date.UTC(y, m - 1, d));
    dt.setUTCDate(dt.getUTCDate() + delta);
    setDate(dt.toISOString().slice(0, 10));
  };

  return (
    <div className="col-span-12 space-y-4 md:space-y-6">
      <div className="rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-white/90">
              Daily App Usage
            </h2>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Zona waktu Asia/Jakarta · unik berdasarkan device_id
            </p>
          </div>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <label className="flex flex-col gap-1 text-sm text-gray-500 dark:text-gray-400">
              Tanggal
              <span className="flex gap-2">
                <button
                  type="button"
                  onClick={() => shiftDay(-1)}
                  className="rounded-lg border border-gray-200 px-3 py-2 dark:border-gray-700"
                  aria-label="Hari sebelumnya"
                >
                  ←
                </button>
                <input
                  type="date"
                  value={date}
                  max={todayWIB()}
                  onChange={(e) => e.target.value && setDate(e.target.value)}
                  className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-gray-800 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90"
                />
                <button
                  type="button"
                  onClick={() => shiftDay(1)}
                  className="rounded-lg border border-gray-200 px-3 py-2 dark:border-gray-700"
                  aria-label="Hari berikutnya"
                >
                  →
                </button>
              </span>
            </label>
            <label className="flex flex-col gap-1 text-sm text-gray-500 dark:text-gray-400">
              Apps name (opsional)
              <span className="flex gap-2">
                <input
                  type="text"
                  value={appsName}
                  onChange={(e) => setAppsName(e.target.value)}
                  placeholder="daebak-music"
                  className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-gray-800 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90"
                />
                <button
                  type="button"
                  onClick={() => setAppliedApps(appsName.trim().toLowerCase())}
                  className="rounded-lg bg-brand-500 px-4 py-2 text-white"
                >
                  Terapkan
                </button>
              </span>
            </label>
          </div>
        </div>

        {error && (
          <p className="mt-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400">
            {error}
          </p>
        )}

        <div className="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">User aktif (DAU)</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : (data?.dau ?? 0).toLocaleString("id-ID")}
            </p>
            <p className="mt-1 text-xs text-gray-400">unik device_id</p>
          </div>
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">Total hits</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : (data?.total_hits ?? 0).toLocaleString("id-ID")}
            </p>
            <p className="mt-1 text-xs text-gray-400">semua tembakan tercatat</p>
          </div>
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">Rata-rata hit/device</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : avgPerDevice}
            </p>
            <p className="mt-1 text-xs text-gray-400">{data?.date ?? date} WIB</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-4 md:gap-6">
        <div className="col-span-12">
          <HourlyChart hourly={data?.hourly ?? emptyHourly} />
        </div>
        <div className="col-span-12">
          <CountriesChart data={data?.by_country ?? []} />
        </div>
        <div className="col-span-12">
          <AppsChart data={data?.by_app ?? []} />
        </div>
      </div>
    </div>
  );
}
