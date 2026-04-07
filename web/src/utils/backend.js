const DEFAULT_API_BASE_URL = "/api/v1";

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL ||
  import.meta.env.VITE_API_URL ||
  DEFAULT_API_BASE_URL;

const normalizeCsvList = (value = "") =>
  String(value || "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);

const stunUrls = normalizeCsvList(
  import.meta.env.VITE_WEBRTC_STUN_URLS || "stun:stun.l.google.com:19302",
);
const turnUrls = normalizeCsvList(import.meta.env.VITE_WEBRTC_TURN_URLS);
const turnUsername = String(
  import.meta.env.VITE_WEBRTC_TURN_USERNAME || "",
).trim();
const turnCredential = String(
  import.meta.env.VITE_WEBRTC_TURN_CREDENTIAL || "",
).trim();

export const WEBRTC_ICE_SERVERS = [
  ...stunUrls.map((urls) => ({ urls })),
  ...turnUrls.map((urls) => ({
    urls,
    username: turnUsername,
    credential: turnCredential,
  })),
];

export const BACKEND_ORIGIN = (() => {
  try {
    const base =
      typeof window !== "undefined" ? window.location.origin : undefined;
    const url = new URL(API_BASE_URL, base);
    return `${url.protocol}//${url.host}`;
  } catch {
    return window.location.origin;
  }
})();

export const getBackendAssetUrl = (path = "") => {
  const value = String(path || "").trim();
  if (!value) return "";
  if (/^https?:\/\//i.test(value)) return value;
  return `${BACKEND_ORIGIN}${value.startsWith("/") ? "" : "/"}${value}`;
};
