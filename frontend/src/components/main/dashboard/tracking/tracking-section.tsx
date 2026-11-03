"use client";

import { api, type DailyStatsResponse } from "@/lib/api";
import { useCallback, useEffect, useMemo, useState } from "react";
import AppsChart from "./apps-chart";
import CountriesChart from "./countries-chart";
import CountryFlag from "./country-flag";
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

export type CountryOption = { code: string; name: string };

export default function TrackingSection() {
  const [date, setDate] = useState<string>(todayWIB());
  const [appliedApps, setAppliedApps] = useState<string[]>([]);
  const [appOptions, setAppOptions] = useState<string[]>([]);
  const [appliedCountries, setAppliedCountries] = useState<string[]>([]);
  const [countryOptions, setCountryOptions] = useState<CountryOption[]>([]);
  // Tanggal asal daftar opsi pill saat ini (bisa beda dengan tanggal applied
  // saat user menjelajah tanggal di dalam modal tanpa Terapkan).
  const [optionsDate, setOptionsDate] = useState<string | null>(null);
  const [data, setData] = useState<DailyStatsResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  // Draft filter di dalam modal.
  const [filterOpen, setFilterOpen] = useState(false);
  const [draftDate, setDraftDate] = useState<string>(date);
  const [draftApps, setDraftApps] = useState<string[]>([]);
  const [draftCountries, setDraftCountries] = useState<string[]>([]);
  const [countryQuery, setCountryQuery] = useState("");
  const [optionsLoading, setOptionsLoading] = useState(false);

  // Sync draft tiap modal dibuka.
  useEffect(() => {
    if (filterOpen) {
      setDraftDate(date);
      setDraftApps(appliedApps);
      setDraftCountries(appliedCountries);
      setCountryQuery("");
    }
  }, [filterOpen, date, appliedApps, appliedCountries]);

  const filteredCountries = useMemo(() => {
    const q = countryQuery.trim().toLowerCase();
    if (!q) return countryOptions;
    return countryOptions.filter(
      (c) => c.name.toLowerCase().includes(q) || c.code.toLowerCase().includes(q),
    );
  }, [countryOptions, countryQuery]);

  // Pill mengikuti tanggal draft: saat tanggal di modal diubah, ambil ulang
  // daftar opsi dari endpoint khusus (scope tanggal saja) dengan debounce.
  // Dibandingkan ke optionsDate (tanggal asal opsi), bukan tanggal applied —
  // kalau tidak, kembali ke tanggal semula justru di-skip dan pill macet kosong.
  useEffect(() => {
    if (!filterOpen) return;
    if (draftDate === optionsDate) return;
    if (!/^\d{4}-\d{2}-\d{2}$/.test(draftDate)) return;
    setOptionsLoading(true);
    const t = setTimeout(async () => {
      try {
        const res = await api.getFilterOptions(draftDate);
        setCountryOptions(
          res.countries.map((c) => ({
            code: c.country_code || c.country,
            name: c.country,
          })),
        );
        setAppOptions(res.apps);
        setOptionsDate(draftDate);
      } catch {
        // Biarkan opsi lama jika gagal; user tetap bisa Terapkan.
      } finally {
        setOptionsLoading(false);
      }
    }, 400);
    return () => clearTimeout(t);
  }, [filterOpen, draftDate, optionsDate]);

  const fetchStats = useCallback(async (d: string, apps: string[], countries: string[]) => {
    setLoading(true);
    setError("");
    try {
      const res = await api.getDailyStats({
        date: d,
        apps_name: apps.length > 0 ? apps.join(",") : undefined,
        country_code: countries.length > 0 ? countries.join(",") : undefined,
      });
      // Ensure 24 buckets even if backend returns fewer.
      const byHour = new Map(res.hourly.map((b) => [b.hour, b]));
      const hourly = Array.from({ length: 24 }, (_, h) => byHour.get(h) ?? emptyHourly[h]);
      setData({ ...res, hourly });
      // Option lists are scoped by date only (backend ignores active filters
      // for them), so pills always match the selected date.
      setCountryOptions(
        (res.all_countries ?? res.by_country).map((c) => ({
          code: c.country_code || c.country,
          name: c.country,
        })),
      );
      setAppOptions(res.all_apps ?? res.by_app.map((a) => a.apps_name));
      setOptionsDate(d);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchStats(date, appliedApps, appliedCountries);
  }, [date, appliedApps, appliedCountries, fetchStats]);

  const avgPerDevice = useMemo(() => {
    if (!data || data.dau === 0) return "0";
    return (data.total_hits / data.dau).toFixed(1);
  }, [data]);

  const shiftDraftDay = (delta: number) => {
    const [y, m, d] = draftDate.split("-").map(Number);
    const dt = new Date(Date.UTC(y, m - 1, d));
    dt.setUTCDate(dt.getUTCDate() + delta);
    setDraftDate(dt.toISOString().slice(0, 10));
  };

  const activeFilterCount =
    appliedApps.length + appliedCountries.length + (date !== todayWIB() ? 1 : 0);

  const scopeNote = useMemo(() => {
    const parts: string[] = [];
    if (appliedApps.length > 0) parts.push(appliedApps.join(", "));
    if (appliedCountries.length > 0) parts.push(appliedCountries.join(", "));
    return parts.length > 0 ? parts.join(" · ") : undefined;
  }, [appliedApps, appliedCountries]);

  const applyFilter = () => {
    setDate(draftDate);
    setAppliedApps(draftApps);
    setAppliedCountries(draftCountries);
    setFilterOpen(false);
  };

  const resetFilter = () => {
    setDraftDate(todayWIB());
    setDraftApps([]);
    setDraftCountries([]);
    setDate(todayWIB());
    setAppliedApps([]);
    setAppliedCountries([]);
    setFilterOpen(false);
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
            <p className="mt-1 text-xs text-gray-400">
              {date} WIB{appliedApps.length > 0 ? ` · ${appliedApps.join(", ")}` : ""}
              {appliedCountries.length > 0 ? ` · ${appliedCountries.join(", ")}` : ""}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setFilterOpen(true)}
            className="flex items-center gap-2 self-start rounded-lg border border-gray-200 px-4 py-2 text-sm font-medium text-gray-700 lg:self-auto dark:border-gray-700 dark:text-gray-200"
          >
            <span>⚙</span> Filter
            {activeFilterCount > 0 && (
              <span className="flex size-5 items-center justify-center rounded-full bg-brand-500 text-[11px] font-semibold text-white">
                {activeFilterCount}
              </span>
            )}
          </button>
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
          <HourlyChart hourly={data?.hourly ?? emptyHourly} scopeNote={scopeNote} />
        </div>
        <div className="col-span-12">
          <CountriesChart data={data?.by_country ?? []} />
        </div>
        <div className="col-span-12">
          <AppsChart data={data?.by_app ?? []} />
        </div>
      </div>

      {filterOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-gray-900/50"
            onClick={() => setFilterOpen(false)}
            aria-hidden
          />
          <div className="relative max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl dark:bg-gray-900">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">Filter</h3>
              <button
                type="button"
                onClick={() => setFilterOpen(false)}
                aria-label="Tutup filter"
                className="flex size-8 items-center justify-center rounded-full bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-300"
              >
                ✕
              </button>
            </div>

            <p className="mb-2 text-sm font-semibold text-gray-800 dark:text-white/90">Tanggal</p>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => shiftDraftDay(-1)}
                aria-label="Hari sebelumnya"
                className="rounded-lg border border-gray-200 px-3 py-2 dark:border-gray-700"
              >
                ←
              </button>
              <input
                type="date"
                value={draftDate}
                max={todayWIB()}
                onChange={(e) => e.target.value && setDraftDate(e.target.value)}
                className="w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-white/90"
              />
              <button
                type="button"
                onClick={() => shiftDraftDay(1)}
                aria-label="Hari berikutnya"
                className="rounded-lg border border-gray-200 px-3 py-2 dark:border-gray-700"
              >
                →
              </button>
            </div>

            <p className="mt-5 mb-2 text-sm font-semibold text-gray-800 dark:text-white/90">
              Apps name <span className="font-normal text-gray-400">(multi-select)</span>
              {optionsLoading && <span className="ml-2 text-xs font-normal text-gray-400">memuat…</span>}
            </p>
            {appOptions.length === 0 ? (
              <p className="text-sm text-gray-500 dark:text-gray-400">Belum ada data aplikasi.</p>
            ) : (
              <div className="flex max-h-56 flex-wrap gap-2 overflow-y-auto">
                {appOptions.map((app) => {
                  const active = draftApps.includes(app);
                  return (
                    <button
                      key={app}
                      type="button"
                      aria-pressed={active}
                      onClick={() =>
                        setDraftApps((prev) =>
                          prev.includes(app) ? prev.filter((x) => x !== app) : [...prev, app],
                        )
                      }
                      title={app}
                      className={
                        active
                          ? "rounded-full bg-brand-500 px-4 py-2 text-xs font-medium text-white"
                          : "rounded-full border border-gray-200 px-4 py-2 text-xs text-gray-600 hover:border-brand-500 hover:text-brand-500 dark:border-gray-700 dark:text-gray-300"
                      }
                    >
                      {active ? "✓ " : ""}{app}
                    </button>
                  );
                })}
              </div>
            )}

            <p className="mt-5 mb-2 text-sm font-semibold text-gray-800 dark:text-white/90">
              Country <span className="font-normal text-gray-400">(multi-select)</span>
              {optionsLoading && <span className="ml-2 text-xs font-normal text-gray-400">memuat…</span>}
            </p>
            <input
              type="text"
              value={countryQuery}
              onChange={(e) => setCountryQuery(e.target.value)}
              placeholder="Cari negara…"
              className="mb-2 w-full rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-white/90"
            />
            {countryOptions.length === 0 ? (
              <p className="text-sm text-gray-500 dark:text-gray-400">Belum ada data negara.</p>
            ) : filteredCountries.length === 0 ? (
              <p className="text-sm text-gray-500 dark:text-gray-400">
                Tidak cocok dengan “{countryQuery}”.
              </p>
            ) : (
              <div className="flex max-h-56 flex-wrap gap-2 overflow-y-auto">
                {filteredCountries.map((c) => {
                  const active = draftCountries.includes(c.code);
                  return (
                    <button
                      key={c.code}
                      type="button"
                      aria-pressed={active}
                      onClick={() =>
                        setDraftCountries((prev) =>
                          prev.includes(c.code)
                            ? prev.filter((x) => x !== c.code)
                            : [...prev, c.code],
                        )
                      }
                      title={`${c.name} (${c.code})`}
                      className={
                        active
                          ? "flex items-center gap-2 rounded-full bg-brand-500 py-1.5 pr-4 pl-1.5 text-xs font-medium text-white"
                          : "flex items-center gap-2 rounded-full border border-gray-200 py-1.5 pr-4 pl-1.5 text-xs text-gray-600 hover:border-brand-500 hover:text-brand-500 dark:border-gray-700 dark:text-gray-300"
                      }
                    >
                      <CountryFlag
                        code={c.code}
                        alt={c.name}
                        imgClassName="h-5 w-5 shrink-0 rounded-full object-cover"
                        badgeClassName="flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-gray-100 text-[10px] font-bold text-gray-500 dark:bg-gray-700 dark:text-gray-300"
                      />
                      {active ? "✓ " : ""}{c.name}
                    </button>
                  );
                })}
              </div>
            )}

            <div className="mt-6 flex gap-2">
              <button
                type="button"
                onClick={resetFilter}
                className="rounded-xl border border-gray-200 px-4 py-2.5 text-sm text-gray-600 dark:border-gray-700 dark:text-gray-300"
              >
                Reset
              </button>
              <button
                type="button"
                onClick={applyFilter}
                className="flex-1 rounded-xl bg-brand-500 px-4 py-2.5 text-sm font-medium text-white"
              >
                Terapkan
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
