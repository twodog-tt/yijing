export function getSiteUrl(): string {
  const url =
    process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";
  return url.replace(/\/$/, "");
}

export function getDivinationUrl(id: number): string {
  return `${getSiteUrl()}/divination/${id}`;
}
