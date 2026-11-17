import createClient from "openapi-fetch";
import type { paths } from "./schema";
export { type components } from "./schema";
export function apiClient(csrf?: string, baseUrl = "") {
  return createClient<paths>({
    baseUrl,
    credentials: "include",
    headers: csrf ? { "X-CSRF-Token": csrf } : undefined,
  });
}
