"use client";

import type { CountryStat } from "@/lib/api";
import { useMemo, useState } from "react";
import CountryMap, { type Marker } from "../ecommerce/demographic-card/country-map";

// Titik tengah kasar tiap kode ISO 3166-1 alpha-2 untuk marker peta.
// Key tambahan berupa nama/alias lama tetap didukung karena backend
// menyimpan country apa adanya (uppercased).
const COUNTRY_COORDS: Record<string, { name: string; latLng: [number, number] }> = {
  ID: { name: "Indonesia", latLng: [-2.5, 118] },
  INDONESIA: { name: "Indonesia", latLng: [-2.5, 118] },
  MY: { name: "Malaysia", latLng: [4.2, 102.2] },
  MALAYSIA: { name: "Malaysia", latLng: [4.2, 102.2] },
  SG: { name: "Singapore", latLng: [1.35, 103.8] },
  SINGAPORE: { name: "Singapore", latLng: [1.35, 103.8] },
  SINGAPURA: { name: "Singapore", latLng: [1.35, 103.8] },
  US: { name: "USA", latLng: [37.09, -95.71] },
  USA: { name: "USA", latLng: [37.09, -95.71] },
  AMERIKA: { name: "USA", latLng: [37.09, -95.71] },
  AMERICA: { name: "USA", latLng: [37.09, -95.71] },
  "UNITED STATES": { name: "USA", latLng: [37.09, -95.71] },
  GB: { name: "United Kingdom", latLng: [55.37, -3.43] },
  UK: { name: "United Kingdom", latLng: [55.37, -3.43] },
  FR: { name: "France", latLng: [46.22, 2.21] },
  DE: { name: "Germany", latLng: [51.16, 10.45] },
  NL: { name: "Netherlands", latLng: [52.13, 5.29] },
  AU: { name: "Australia", latLng: [-25.27, 133.77] },
  JP: { name: "Japan", latLng: [36.2, 138.25] },
  KR: { name: "South Korea", latLng: [35.9, 127.76] },
  IN: { name: "India", latLng: [20.59, 78.96] },
  PH: { name: "Philippines", latLng: [12.87, 121.77] },
  TH: { name: "Thailand", latLng: [15.87, 100.99] },
  VN: { name: "Vietnam", latLng: [14.05, 108.27] },
  YE: { name: "Yemen", latLng: [15.55, 48.51] },
  SA: { name: "Saudi Arabia", latLng: [24.0, 45.0] },
  AE: { name: "UAE", latLng: [24.45, 54.37] },
  AF: { name: "Afghanistan", latLng: [33.93, 67.7] },
  AL: { name: "Albania", latLng: [41.15, 20.16] },
  DZ: { name: "Algeria", latLng: [28.03, 1.65] },
  AD: { name: "Andorra", latLng: [42.54, 1.58] },
  AO: { name: "Angola", latLng: [-11.2, 17.87] },
  AR: { name: "Argentina", latLng: [-38.41, -63.61] },
  AM: { name: "Armenia", latLng: [40.06, 45.03] },
  AT: { name: "Austria", latLng: [47.51, 14.55] },
  AZ: { name: "Azerbaijan", latLng: [40.14, 47.57] },
  BS: { name: "Bahamas", latLng: [25.03, -77.39] },
  BH: { name: "Bahrain", latLng: [26.02, 50.55] },
  BD: { name: "Bangladesh", latLng: [23.68, 90.35] },
  BB: { name: "Barbados", latLng: [13.19, -59.54] },
  BY: { name: "Belarus", latLng: [53.7, 27.95] },
  BE: { name: "Belgium", latLng: [50.85, 4.35] },
  BZ: { name: "Belize", latLng: [17.18, -88.3] },
  BJ: { name: "Benin", latLng: [9.3, 2.31] },
  BT: { name: "Bhutan", latLng: [27.51, 90.43] },
  BO: { name: "Bolivia", latLng: [-16.29, -63.58] },
  BA: { name: "Bosnia and Herzegovina", latLng: [43.91, 17.67] },
  BW: { name: "Botswana", latLng: [-22.32, 23.94] },
  BR: { name: "Brazil", latLng: [-14.23, -51.92] },
  BN: { name: "Brunei", latLng: [35.67, 53.93] },
  BG: { name: "Bulgaria", latLng: [42.73, 25.48] },
  BF: { name: "Burkina Faso", latLng: [12.23, -1.56] },
  BI: { name: "Burundi", latLng: [-3.37, 29.91] },
  KH: { name: "Cambodia", latLng: [12.56, 104.99] },
  CM: { name: "Cameroon", latLng: [7.36, 12.35] },
  CA: { name: "Canada", latLng: [56.13, -106.34] },
  CV: { name: "Cape Verde", latLng: [16.0, -24.0] },
  CF: { name: "Central African Republic", latLng: [6.61, 20.93] },
  TD: { name: "Chad", latLng: [15.45, 18.73] },
  CL: { name: "Chile", latLng: [-35.67, -71.54] },
  CN: { name: "China", latLng: [35.86, 104.19] },
  CO: { name: "Colombia", latLng: [4.57, -74.29] },
  KM: { name: "Comoros", latLng: [-11.87, 43.87] },
  CG: { name: "Congo", latLng: [-0.22, 15.82] },
  CD: { name: "DR Congo", latLng: [-4.03, 21.75] },
  CR: { name: "Costa Rica", latLng: [9.74, -83.75] },
  CI: { name: "Ivory Coast", latLng: [7.53, -5.54] },
  HR: { name: "Croatia", latLng: [45.1, 15.2] },
  CU: { name: "Cuba", latLng: [21.52, -77.78] },
  CY: { name: "Cyprus", latLng: [35.12, 33.42] },
  CZ: { name: "Czechia", latLng: [49.81, 15.47] },
  DK: { name: "Denmark", latLng: [56.26, 9.5] },
  DJ: { name: "Djibouti", latLng: [11.82, 42.59] },
  DM: { name: "Dominica", latLng: [15.41, -61.37] },
  DO: { name: "Dominican Republic", latLng: [18.73, -70.16] },
  EC: { name: "Ecuador", latLng: [-1.83, -78.18] },
  EG: { name: "Egypt", latLng: [26.82, 30.8] },
  SV: { name: "El Salvador", latLng: [13.79, -88.89] },
  GQ: { name: "Equatorial Guinea", latLng: [1.65, 10.26] },
  ER: { name: "Eritrea", latLng: [15.17, 39.78] },
  EE: { name: "Estonia", latLng: [58.59, 25.0] },
  ET: { name: "Ethiopia", latLng: [9.14, 40.48] },
  FJ: { name: "Fiji", latLng: [-16.57, 179.41] },
  FI: { name: "Finland", latLng: [61.92, 25.74] },
  GA: { name: "Gabon", latLng: [-0.8, 11.61] },
  GM: { name: "Gambia", latLng: [13.44, -16.57] },
  GE: { name: "Georgia", latLng: [42.31, 43.36] },
  GH: { name: "Ghana", latLng: [7.94, -1.02] },
  GR: { name: "Greece", latLng: [39.07, 21.82] },
  GD: { name: "Grenada", latLng: [12.26, -61.6] },
  GT: { name: "Guatemala", latLng: [15.78, -90.23] },
  GN: { name: "Guinea", latLng: [9.94, -9.69] },
  GW: { name: "Guinea-Bissau", latLng: [11.8, -15.18] },
  GY: { name: "Guyana", latLng: [4.86, -58.93] },
  HT: { name: "Haiti", latLng: [18.97, -72.28] },
  HN: { name: "Honduras", latLng: [15.19, -86.24] },
  HK: { name: "Hong Kong", latLng: [22.31, 114.16] },
  HU: { name: "Hungary", latLng: [47.16, 19.5] },
  IS: { name: "Iceland", latLng: [65.0, -18.09] },
  IR: { name: "Iran", latLng: [32.42, 53.68] },
  IQ: { name: "Iraq", latLng: [33.0, 44.36] },
  IE: { name: "Ireland", latLng: [53.41, -8.24] },
  IL: { name: "Israel", latLng: [30.37, 34.35] },
  IT: { name: "Italy", latLng: [41.87, 12.56] },
  JM: { name: "Jamaica", latLng: [18.1, -77.29] },
  JO: { name: "Jordan", latLng: [30.58, 36.24] },
  KZ: { name: "Kazakhstan", latLng: [48.01, 66.92] },
  KE: { name: "Kenya", latLng: [-0.02, 37.86] },
  KI: { name: "Kiribati", latLng: [1.87, -157.36] },
  KP: { name: "North Korea", latLng: [40.33, 127.51] },
  KW: { name: "Kuwait", latLng: [29.31, 47.48] },
  KG: { name: "Kyrgyzstan", latLng: [41.2, 74.76] },
  LA: { name: "Laos", latLng: [19.85, 102.49] },
  LV: { name: "Latvia", latLng: [56.87, 24.6] },
  LB: { name: "Lebanon", latLng: [33.85, 35.86] },
  LS: { name: "Lesotho", latLng: [-29.61, 28.23] },
  LR: { name: "Liberia", latLng: [6.42, -9.42] },
  LY: { name: "Libya", latLng: [26.33, 17.22] },
  LI: { name: "Liechtenstein", latLng: [47.16, 9.55] },
  LT: { name: "Lithuania", latLng: [55.16, 23.88] },
  LU: { name: "Luxembourg", latLng: [49.81, 6.12] },
  MO: { name: "Macau", latLng: [22.19, 113.54] },
  MK: { name: "North Macedonia", latLng: [41.6, 21.74] },
  MG: { name: "Madagascar", latLng: [-18.76, 46.86] },
  MW: { name: "Malawi", latLng: [-13.25, 34.3] },
  MV: { name: "Maldives", latLng: [3.2, 73.22] },
  ML: { name: "Mali", latLng: [17.57, -3.99] },
  MT: { name: "Malta", latLng: [35.93, 14.37] },
  MH: { name: "Marshall Islands", latLng: [7.13, 171.18] },
  MR: { name: "Mauritania", latLng: [21.0, -10.94] },
  MU: { name: "Mauritius", latLng: [-20.34, 57.55] },
  MX: { name: "Mexico", latLng: [23.63, -102.55] },
  FM: { name: "Micronesia", latLng: [7.42, 150.55] },
  MD: { name: "Moldova", latLng: [47.41, 28.36] },
  MC: { name: "Monaco", latLng: [43.73, 7.41] },
  MN: { name: "Mongolia", latLng: [46.86, 103.84] },
  ME: { name: "Montenegro", latLng: [43.27, 19.37] },
  MA: { name: "Morocco", latLng: [31.79, -7.09] },
  MZ: { name: "Mozambique", latLng: [-18.66, 35.52] },
  MM: { name: "Myanmar", latLng: [21.91, 95.95] },
  NA: { name: "Namibia", latLng: [-22.95, 18.49] },
  NR: { name: "Nauru", latLng: [-0.52, 166.93] },
  NP: { name: "Nepal", latLng: [28.39, 84.12] },
  NC: { name: "New Caledonia", latLng: [-20.9, 165.61] },
  NZ: { name: "New Zealand", latLng: [-40.9, 174.88] },
  NI: { name: "Nicaragua", latLng: [12.86, -85.2] },
  NE: { name: "Niger", latLng: [17.6, 8.08] },
  NG: { name: "Nigeria", latLng: [9.08, 8.67] },
  NO: { name: "Norway", latLng: [64.14, 11.52] },
  OM: { name: "Oman", latLng: [21.51, 55.92] },
  PK: { name: "Pakistan", latLng: [30.37, 69.34] },
  PW: { name: "Palau", latLng: [7.51, 134.58] },
  PS: { name: "Palestine", latLng: [31.95, 35.23] },
  PA: { name: "Panama", latLng: [8.53, -80.78] },
  PG: { name: "Papua New Guinea", latLng: [-6.31, 147.0] },
  PY: { name: "Paraguay", latLng: [-23.44, -58.44] },
  PE: { name: "Peru", latLng: [-9.18, -75.0] },
  PL: { name: "Poland", latLng: [51.91, 19.14] },
  PT: { name: "Portugal", latLng: [39.39, -8.22] },
  PR: { name: "Puerto Rico", latLng: [18.22, -66.59] },
  QA: { name: "Qatar", latLng: [25.35, 51.18] },
  RO: { name: "Romania", latLng: [45.94, 24.96] },
  RU: { name: "Russia", latLng: [61.52, 90.31] },
  RW: { name: "Rwanda", latLng: [-1.94, 29.87] },
  KN: { name: "Saint Kitts and Nevis", latLng: [17.35, -62.78] },
  LC: { name: "Saint Lucia", latLng: [13.9, -60.97] },
  VC: { name: "Saint Vincent", latLng: [12.98, -61.28] },
  WS: { name: "Samoa", latLng: [-13.75, -172.1] },
  SM: { name: "San Marino", latLng: [43.94, 12.45] },
  ST: { name: "Sao Tome and Principe", latLng: [0.18, 6.61] },
  SN: { name: "Senegal", latLng: [14.49, -14.45] },
  RS: { name: "Serbia", latLng: [44.01, 20.26] },
  SC: { name: "Seychelles", latLng: [-4.67, 55.49] },
  SL: { name: "Sierra Leone", latLng: [8.46, -11.77] },
  SK: { name: "Slovakia", latLng: [48.66, 19.69] },
  SI: { name: "Slovenia", latLng: [46.15, 14.99] },
  SB: { name: "Solomon Islands", latLng: [-9.64, 160.15] },
  SO: { name: "Somalia", latLng: [5.15, 46.19] },
  ZA: { name: "South Africa", latLng: [-30.55, 22.93] },
  SS: { name: "South Sudan", latLng: [6.87, 31.3] },
  ES: { name: "Spain", latLng: [40.46, -3.74] },
  LK: { name: "Sri Lanka", latLng: [7.87, 80.77] },
  SD: { name: "Sudan", latLng: [12.86, 30.21] },
  SR: { name: "Suriname", latLng: [3.91, -56.02] },
  SZ: { name: "Eswatini", latLng: [-26.52, 31.46] },
  SE: { name: "Sweden", latLng: [60.12, 18.64] },
  CH: { name: "Switzerland", latLng: [46.81, 8.22] },
  SY: { name: "Syria", latLng: [34.8, 38.99] },
  TW: { name: "Taiwan", latLng: [23.69, 120.96] },
  TJ: { name: "Tajikistan", latLng: [38.86, 71.27] },
  TZ: { name: "Tanzania", latLng: [-6.36, 34.88] },
  TL: { name: "Timor-Leste", latLng: [-8.87, 125.72] },
  TG: { name: "Togo", latLng: [8.61, 0.82] },
  TO: { name: "Tonga", latLng: [-21.17, -175.2] },
  TT: { name: "Trinidad and Tobago", latLng: [10.69, -61.22] },
  TN: { name: "Tunisia", latLng: [33.88, 9.53] },
  TR: { name: "Turkey", latLng: [38.96, 35.24] },
  TM: { name: "Turkmenistan", latLng: [39.3, 59.32] },
  TV: { name: "Tuvalu", latLng: [-7.47, 178.06] },
  UG: { name: "Uganda", latLng: [1.37, 32.29] },
  UA: { name: "Ukraine", latLng: [48.37, 31.16] },
  UKRAINA: { name: "Ukraine", latLng: [48.37, 31.16] },
  UY: { name: "Uruguay", latLng: [-32.52, -55.76] },
  UZ: { name: "Uzbekistan", latLng: [41.37, 64.58] },
  VU: { name: "Vanuatu", latLng: [-15.37, 167.0] },
  VE: { name: "Venezuela", latLng: [6.42, -66.58] },
  ZM: { name: "Zambia", latLng: [-13.13, 27.84] },
  ZW: { name: "Zimbabwe", latLng: [-19.01, 29.15] },
};

function normalizeCountry(code: string): string {
  return code.trim().toUpperCase();
}

function codeForMarker(d: { country: string; country_code?: string }): string {
  const cc = (d.country_code ?? "").trim();
  if (cc) return cc;
  return d.country;
}

function countryLabel(code: string, fallback?: string): string {
  return COUNTRY_COORDS[normalizeCountry(code)]?.name ?? fallback ?? code;
}

function flagCode(code: string): string | null {
  const c = normalizeCountry(code);
  // flagcdn hanya dukung ISO alpha-2
  if (/^[A-Z]{2}$/.test(c)) return c.toLowerCase();
  return null;
}

function CountryFlag({ code, alt }: { code: string; alt: string }) {
  const [failed, setFailed] = useState(false);
  const fc = flagCode(code);
  if (!fc || failed) {
    return (
      <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-xs font-bold text-brand-500 dark:bg-brand-500/10">
        {code.slice(0, 2).toUpperCase()}
      </div>
    );
  }
  return (
    <img
      src={`https://flagcdn.com/w80/${fc}.png`}
      srcSet={`https://flagcdn.com/w80/${fc}.png 1x, https://flagcdn.com/w160/${fc}.png 2x`}
      alt={alt}
      loading="lazy"
      onError={() => setFailed(true)}
      className="size-8 shrink-0 rounded-lg object-cover"
    />
  );
}

export default function CountriesChart({ data }: { data: CountryStat[] }) {
  const totalUnique = useMemo(
    () => data.reduce((s, d) => s + d.unique_devices, 0),
    [data],
  );

  const markers: Marker[] = useMemo(
    () =>
      data
        .filter((d) => COUNTRY_COORDS[normalizeCountry(codeForMarker(d))])
        .map((d) => {
          const code = codeForMarker(d);
          return {
            latLng: COUNTRY_COORDS[normalizeCountry(code)].latLng,
            name: `${countryLabel(code, d.country)}: ${d.unique_devices}`,
            style: { fill: "#465FFF", strokeWidth: 1, stroke: "white" },
          };
        }),
    [data],
  );

  if (data.length === 0) {
    return (
      <div className="h-full rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
        <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">
          Negara paling aktif
        </h3>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          User unik per hari berdasarkan country
        </p>
        <p className="mt-6 text-sm text-gray-500 dark:text-gray-400">
          Belum ada data untuk tanggal ini.
        </p>
      </div>
    );
  }

  return (
    <div className="h-full rounded-2xl border border-gray-200 bg-white p-5 sm:p-6 dark:border-gray-800 dark:bg-white/3">
      <div className="flex justify-between">
        <div>
          <h3 className="text-lg font-semibold text-gray-800 dark:text-white/90">
            Negara paling aktif
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            User unik per hari berdasarkan country
          </p>
        </div>
      </div>

      <div className="my-6 overflow-hidden rounded-2xl border border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-900">
        <div className="h-[260px] w-full sm:h-[300px] lg:h-[360px]">
          <CountryMap markers={markers} />
        </div>
      </div>

      <div className="flex flex-col gap-5">
        {data.slice(0, 5).map((d) => {
          const pct = totalUnique > 0 ? (d.unique_devices / totalUnique) * 100 : 0;
          const code = codeForMarker(d);
          return (
            <div key={`${d.country_code || d.country}`} className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <CountryFlag code={d.country_code || d.country} alt={countryLabel(code, d.country)} />
                <div>
                  <p className="text-sm font-semibold text-gray-800 dark:text-white/90">
                    {countryLabel(code, d.country)}
                  </p>
                  <span className="block text-xs text-gray-500 dark:text-gray-400">
                    {d.unique_devices.toLocaleString("id-ID")} user unik · {d.hits.toLocaleString("id-ID")} hits
                  </span>
                </div>
              </div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">
                {pct.toFixed(2)}%
              </p>
            </div>
          );
        })}
      </div>
    </div>
  );
}
