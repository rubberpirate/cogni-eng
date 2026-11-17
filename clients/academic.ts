import type { components } from "./schema";
export type A<N extends keyof components["schemas"]> = components["schemas"][N];
type Lists = {
  courses: A<"AcademicCourse">;
  topics: A<"AcademicTopic">;
  questions: A<"AcademicQuestion">;
  imports: A<"AcademicImport">;
  papers: A<"AcademicPaperRecord">;
  assignments: A<"AcademicAssigned">;
  layouts: A<"AcademicSheetLayout">;
  batches: A<"AcademicBatch">;
  "scan-sheets": A<"AcademicScanSheet">;
  jobs: A<"AcademicJob">;
};
type Details = {
  imports: A<"AcademicImport">;
  papers: A<"AcademicPaperRecord">;
  sittings: A<"AcademicSitting">;
  jobs: A<"AcademicJob">;
};
// DTOs are generated from the versioned OpenAPI contract. Binary multipart uses the same session and CSRF boundary.
export function educationClient(org: string, csrf: string) {
  const base = `/api/v1/institutions/${encodeURIComponent(org)}/education/`;
  async function request<T>(path: string, body?: unknown): Promise<T> {
    const form = body instanceof FormData;
    const response = await fetch(base + path, {
      method: body === undefined ? "GET" : "POST",
      credentials: "include",
      cache: "no-store",
      headers:
        body === undefined
          ? undefined
          : {
              "X-CSRF-Token": csrf,
              ...(!form ? { "Content-Type": "application/json" } : {}),
            },
      body: body === undefined ? undefined : form ? body : JSON.stringify(body),
      signal: AbortSignal.timeout(form ? 120000 : 15000),
    });
    const data: unknown = await response.json();
    if (!response.ok) {
      const error = data as { message?: string; request_id?: string };
      throw new Error(
        `${error.message ?? "Operation failed"}${error.request_id ? ` (reference ${error.request_id})` : ""}`,
      );
    }
    return data as T;
  }
  return {
    url: (path: string) => base + path,
    list: <K extends keyof Lists>(
      kind: K,
      query: Record<string, string> = {},
    ) => request<Lists[K][]>(kind + "?" + new URLSearchParams(query)),
    detail: <K extends keyof Details>(kind: K, id: string) =>
      request<Details[K]>(kind + "/" + id),
    topics: (body: A<"AcademicTopic">) =>
      request<A<"AcademicTopic">>("topics", body),
    question: (body: A<"AcademicQuestionWrite">, id?: string) =>
      request<A<"AcademicQuestionCreated"> | A<"AcademicQuestion">>(
        "questions" + (id ? "/" + id : ""),
        body,
      ),
    review: (id: string, body: A<"AcademicQuestionReview">) =>
      request("questions/" + id + "/review", body),
    uploadImport: (body: FormData) =>
      request<A<"AcademicImport">>("imports/stage/upload", body),
    importChange: (
      id: string,
      action: "edit" | "commit",
      body: A<"AcademicImportChange">,
    ) => request<A<"AcademicImport">>(`imports/${id}/${action}`, body),
    parseDocument: (id: string, body: A<"AcademicParseDocument">) =>
      request<A<"AcademicImport">>(`imports/${id}/parse`, body),
    evidence: (id: string) =>
      request<A<"AcademicTopicEvidence">[]>(`evidence/${id}`),
    history: (id: string) =>
      request<A<"AcademicResultHistory">[]>(`result-history/${id}`),
    thread: (sitting: string, question: string) =>
      request<A<"AcademicThread">>(
        `tutor?${new URLSearchParams({ sitting, question })}`,
      ),
    ask: (body: A<"AcademicTutorAsk">) =>
      request<A<"AcademicThread">>("tutor", body),
    compose: (body: A<"AcademicCompose">) =>
      request<A<"AcademicPaperRecord">>("papers", body),
    publish: (id: string, body: A<"AcademicVersion">) =>
      request(`papers/${id}/publish`, body),
    recipients: (id: string) =>
      request<A<"AcademicRecipient">[]>("recipients/" + id),
    assign: (body: A<"AcademicAssign">) =>
      request<string[]>("assignments", body),
    start: (id: string, takeover = false) =>
      request<A<"AcademicSitting">>(`assignments/${id}/start`, { takeover }),
    save: (id: string, body: A<"AcademicSave">) =>
      request<A<"AcademicSitting">>(`sittings/${id}/save`, body),
    submit: (id: string) =>
      request<A<"AcademicSitting">>(`sittings/${id}/submit`, {}),
    release: (id: string) => request(`sittings/${id}/release`, {}),
    mark: (id: string, body: A<"AcademicReviewMarks">) =>
      request<A<"AcademicSitting">>(`sittings/${id}/review`, body),
    layout: (body: A<"AcademicNewLayout">) =>
      request<A<"AcademicSheetLayout">>("layouts", body),
    batch: (body: A<"AcademicNewBatch">) =>
      request<A<"AcademicBatch">>("batches", body),
    scanUpload: (id: string, body: FormData) =>
      request<A<"AcademicJob">>(`batches/${id}/upload`, body),
    scanReview: (id: string, body: A<"AcademicScanReview">) =>
      request<A<"AcademicSaved">>(`scan-sheets/${id}/review`, body),
    generate: (body: A<"AcademicGenerate">) =>
      request<A<"AcademicJob">>("generate", body),
    jobAction: (id: string, action: "retry" | "cancel") =>
      request(`jobs/${id}/${action}`, {}),
    asset: (body: FormData) => request<A<"AcademicAsset">>("assets", body),
  };
}
export type EducationClient = ReturnType<typeof educationClient>;
