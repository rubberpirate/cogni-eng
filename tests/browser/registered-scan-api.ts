import { expect, type Page } from "@playwright/test";
import { omrFixture } from "./omr-fixtures";
// Retain backward-compatibility coverage of registered evidence without exposing
// the retired scanner as another faculty UI option.
export async function registeredScanAPI(
  page: Page,
  paper: string,
  layout: string,
) {
  const base = `/api/v1/institutions/${omrFixture.org}/education/`;
  const me = await (await page.request.get("/api/v1/me")).json();
  const headers = { "X-CSRF-Token": me.csrf, Origin: "http://localhost:3000" };
  const batchResponse = await page.request.post(base + "batches", {
    headers,
    data: {
      paper_id: paper,
      layout_id: layout,
      sitting_at: new Date().toISOString(),
    },
  });
  expect(batchResponse.status()).toBe(200);
  const batch = await batchResponse.json();
  return {
    base,
    batch,
    async upload(name: string, mimeType: string, buffer: Buffer) {
      const r = await page.request.post(base + `batches/${batch.id}/upload`, {
        headers,
        multipart: { file: { name, mimeType, buffer } },
      });
      expect(r.status()).toBe(200);
      return await r.json();
    },
    async sheets(count = 1) {
      let rows: any[] = [];
      await expect
        .poll(
          async () => {
            rows = await (
              await page.request.get(base + "scan-sheets?batch=" + batch.id)
            ).json();
            return rows.length;
          },
          { timeout: 30000 },
        )
        .toBe(count);
      return rows;
    },
    async review(
      sheet: any,
      assignment: string,
      corrections: Record<string, unknown>,
      status = 200,
    ) {
      const r = await page.request.post(
        base + `scan-sheets/${sheet.id}/review`,
        {
          headers,
          data: {
            assignment_id: assignment,
            reason:
              "Verified original synthetic registered-sheet evidence and identity",
            identity_confirmed: true,
            version: sheet.version,
            corrections,
            commit: true,
          },
        },
      );
      const b = await r.json();
      expect(r.status(), JSON.stringify(b.message ?? "")).toBe(status);
      return b;
    },
  };
}
