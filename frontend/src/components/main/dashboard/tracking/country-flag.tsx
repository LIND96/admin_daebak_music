"use client";

import { useState } from "react";

// flagcdn hanya dukung ISO alpha-2. Return null untuk kode lama/non-ISO
// (mis. "UKR", "UNKNOWN") agar pemanggil bisa fallback ke badge teks.
export function flagCode(code: string): string | null {
  const c = code.trim().toUpperCase();
  if (/^[A-Z]{2}$/.test(c)) return c.toLowerCase();
  return null;
}

export function flagImg(code2: string, width = 80): string {
  return `https://flagcdn.com/w${width}/${code2}.png`;
}

export default function CountryFlag({
  code,
  alt,
  imgClassName = "size-8 shrink-0 rounded-lg object-cover",
  badgeClassName = "flex size-8 shrink-0 items-center justify-center rounded-lg bg-brand-50 text-xs font-bold text-brand-500 dark:bg-brand-500/10",
}: {
  code: string;
  alt: string;
  imgClassName?: string;
  badgeClassName?: string;
}) {
  const [failed, setFailed] = useState(false);
  const fc = flagCode(code);
  if (!fc || failed) {
    return <div className={badgeClassName}>{code.slice(0, 2).toUpperCase()}</div>;
  }
  return (
    <img
      src={flagImg(fc)}
      srcSet={`${flagImg(fc)} 1x, ${flagImg(fc, 160)} 2x`}
      alt={alt}
      loading="lazy"
      onError={() => setFailed(true)}
      className={imgClassName}
    />
  );
}
