"use client";

import { api, type ChatStatsResponse } from "@/lib/api";
import { useEffect, useState } from "react";
import HourlyChart from "./hourly-chart";

/**
 * AI-chat relay monitoring (POST /api/chat/stats/daily).
 * Answers the week-one ops questions: volume per day, unique devices,
 * failure share (fallback-to-direct pressure), token usage, and a rough
 * cost estimate. Mirrors the tracking section layout, intentionally
 * simpler (no country split, no filter modal).
 */

// Rough estimate only — update to match the active provider/model pricing.
const USD_PER_1M_TOKENS_IN = 0.3;
const USD_PER_1M_TOKENS_OUT = 2.5;

function todayWIB(): string {
  const fmt = new Intl.DateTimeFormat("en-CA", {
    timeZone: "Asia/Jakarta",
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
  });
  return fmt.format(new Date());
}

function shiftDay(date: string, delta: number): string {
  const [y, m, d] = date.split("-").map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  dt.setUTCDate(dt.getUTCDate() + delta);
  return dt.toISOString().slice(0, 10);
}

const emptyHourly = Array.from({ length: 24 }, (_, h) => ({
  hour: h,
  hits: 0,
  unique_devices: 0,
}));

export default function ChatSection() {
  const [date, setDate] = useState<string>(todayWIB());
  const [appFilter, setAppFilter] = useState<string>("");
  const [appliedApp, setAppliedApp] = useState<string>("");
  const [data, setData] = useState<ChatStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");

  useEffect(() => {
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError("");
      try {
        const res = await api.getChatStats({
          date: date,
          apps_name: appliedApp.trim() ? appliedApp.trim().toLowerCase() : undefined,
        });
        if (cancelled) return;
        const byHour = new Map(res.hourly.map((b) => [b.hour, b]));
        const hourly = Array.from(
          { length: 24 },
          (_, h) => byHour.get(h) ?? emptyHourly[h],
        );
        setData({ ...res, hourly });
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Gagal memuat data");
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [date, appliedApp]);

  const total = data?.total_messages ?? 0;
  const failed = data?.failed_count ?? 0;
  const failRate = total === 0 ? "0%" : `${((failed / total) * 100).toFixed(1)}%`;
  const tokensIn = data?.tokens_in ?? 0;
  const tokensOut = data?.tokens_out ?? 0;
  const estUsd =
    (tokensIn / 1_000_000) * USD_PER_1M_TOKENS_IN +
    (tokensOut / 1_000_000) * USD_PER_1M_TOKENS_OUT;

  return (
    <div className="col-span-12 space-y-4 md:space-y-6">
      <div className="rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h2 className="text-xl font-semibold text-gray-800 dark:text-white/90">
              AI Chat Relay
            </h2>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
              Zona waktu Asia/Jakarta · kuota per device per hari
            </p>
            <p className="mt-1 text-xs text-gray-400">
              {date} WIB{appliedApp ? ` · ${appliedApp}` : ""}
            </p>
          </div>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <input
              type="text"
              value={appFilter}
              onChange={(e) => setAppFilter(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") setAppliedApp(appFilter);
              }}
              placeholder="Filter apps_name…"
              className="rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-white/90"
            />
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setDate((d) => shiftDay(d, -1))}
                aria-label="Hari sebelumnya"
                className="rounded-lg border border-gray-200 px-3 py-2 dark:border-gray-700"
              >
                ←
              </button>
              <button
                type="button"
                onClick={() => {
                  setDate(todayWIB());
                  setAppFilter("");
                  setAppliedApp("");
                }}
                className="rounded-lg border border-gray-200 px-3 py-2 text-sm dark:border-gray-700"
              >
                Hari ini
              </button>
              <button
                type="button"
                onClick={() => setDate((d) => shiftDay(d, 1))}
                aria-label="Hari berikutnya"
                disabled={date >= todayWIB()}
                className="rounded-lg border border-gray-200 px-3 py-2 disabled:opacity-40 dark:border-gray-700"
              >
                →
              </button>
            </div>
          </div>
        </div>

        {error && (
          <p className="mt-4 rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-400">
            {error}
          </p>
        )}

        <div className="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">Pesan relay</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : total.toLocaleString("id-ID")}
            </p>
            <p className="mt-1 text-xs text-gray-400">
              {loading ? "…" : `${(data?.unique_devices ?? 0).toLocaleString("id-ID")} device unik`}
            </p>
          </div>
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">Gagal provider</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : `${failed.toLocaleString("id-ID")} (${failRate})`}
            </p>
            <p className="mt-1 text-xs text-gray-400">jatuh ke fallback direct</p>
          </div>
          <div className="rounded-xl border border-gray-200 p-4 dark:border-gray-800">
            <p className="text-sm text-gray-500 dark:text-gray-400">Token in / out</p>
            <p className="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
              {loading ? "…" : `${tokensIn.toLocaleString("id-ID")} / ${tokensOut.toLocaleString("id-ID")}`}
            </p>
            <p className="mt-1 text-xs text-gray-400">
              {loading ? "…" : `≈ $${estUsd.toFixed(2)} (estimasi kasar)`}
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-12 gap-4 md:gap-6">
        <div className="col-span-12">
          <HourlyChart hourly={data?.hourly ?? emptyHourly} scopeNote={appliedApp || undefined} />
        </div>
        <div className="col-span-12 lg:col-span-6">
          <div className="rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
            <h3 className="text-base font-semibold text-gray-800 dark:text-white/90">
              Per aplikasi
            </h3>
            {(data?.by_app ?? []).length === 0 ? (
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">Belum ada data.</p>
            ) : (
              <ul className="mt-2 divide-y divide-gray-100 dark:divide-gray-800">
                {data!.by_app.map((a) => (
                  <li key={a.apps_name} className="flex items-center justify-between py-2 text-sm">
                    <span className="text-gray-700 dark:text-gray-200">{a.apps_name}</span>
                    <span className="text-gray-500 dark:text-gray-400">
                      {a.hits.toLocaleString("id-ID")} pesan · {a.unique_devices.toLocaleString("id-ID")} device
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
        <div className="col-span-12 lg:col-span-6">
          <div className="rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
            <h3 className="text-base font-semibold text-gray-800 dark:text-white/90">
              Idol teratas
            </h3>
            {(data?.by_idol ?? []).length === 0 ? (
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">Belum ada data.</p>
            ) : (
              <ul className="mt-2 divide-y divide-gray-100 dark:divide-gray-800">
                {data!.by_idol.map((r) => (
                  <li key={r.idol_name} className="flex items-center justify-between py-2 text-sm">
                    <span className="text-gray-700 dark:text-gray-200">{r.idol_name}</span>
                    <span className="text-gray-500 dark:text-gray-400">
                      {r.hits.toLocaleString("id-ID")} pesan · {r.unique_devices.toLocaleString("id-ID")} device
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
