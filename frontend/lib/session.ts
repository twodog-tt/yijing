const STORAGE_KEY = "yijing_session_key";

function generateUUID(): string {
  if (typeof crypto !== "undefined" && crypto.randomUUID) {
    return crypto.randomUUID();
  }
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/** 仅在浏览器环境调用 */
export function getSessionKey(): string {
  if (typeof window === "undefined") {
    return "";
  }
  let key = localStorage.getItem(STORAGE_KEY);
  if (!key) {
    key = generateUUID();
    localStorage.setItem(STORAGE_KEY, key);
  }
  return key;
}

export function setSessionKey(key: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(STORAGE_KEY, key);
}
