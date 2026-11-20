"use client";
import { useEffect, useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { FormField as Field } from "./form-field";
import { OMRTemplateBuilder } from "./omr-template-builder";
import { referenceFile, type OMRProfile } from "../../../clients/omr-template";
type Paper = A<"AcademicPaperRecord">;
type Check = A<"AcademicOMRCheck">;
type Sheet = A<"AcademicOMRCheckedSheet">;
const questions = (paper: Paper) =>
  paper.spec.sections.flatMap((s) => s.questions);

export function OMRCheckerReader({
  api,
  papers,
  paper,
  onPaper,
}: {
  api: EducationClient;
  papers: Paper[];
  paper: Paper | null;
  onPaper: (paper: Paper | null) => void;
}) {
  const [templates, setTemplates] = useState<A<"AcademicOMRSavedTemplate">[]>(
      [],
    ),
    [template, setTemplate] = useState("");
  const [loading, setLoading] = useState(false),
    [busy, setBusy] = useState(false),
    [saving, setSaving] = useState(false);
  const [files, setFiles] = useState<File[]>([]),
    [sample, setSample] = useState<File | null>(null);
  const [failed, setFailed] = useState<{ file: File; error: string }[]>([]),
    [error, setError] = useState(""),
    [notice, setNotice] = useState("");
  const [revision, setRevision] = useState(0),
    [uploading, setUploading] = useState("");
  const paperID = paper?.state === "published" ? paper.id : "";
  useEffect(() => {
    let active = true;
    setTemplates([]);
    setTemplate("");
    setFiles([]);
    setFailed([]);
    setSample(null);
    setNotice("");
    setError("");
    if (!paperID) {
      setLoading(false);
      return;
    }
    setLoading(true);
    void api
      .list("omr-templates", { paper: paperID })
      .then((items) => {
        if (!active) return;
        setTemplates(items);
        let previous = "";
        try {
          previous = localStorage.getItem("omr-template:" + paperID) ?? "";
        } catch {
          /* Template storage is on the server; browser preferences are optional. */
        }
        if (items.some((t) => t.id === previous)) setTemplate(previous);
      })
      .catch((e) => {
        if (active) setError(e.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [api, paperID]);
  function selectTemplate(id: string) {
    setTemplate(id);
    if (paperID)
      try {
        localStorage.setItem("omr-template:" + paperID, id);
      } catch {
        /* A browser preference failure must not undo a server save. */
      }
  }
  async function upload(selected: File[]) {
    if (!paperID || !template || !selected.length) return;
    if (selected.length > 20) {
      setError(
        "Choose up to 20 files per upload. A PDF can contain up to 20 sheets.",
      );
      return;
    }
    setBusy(true);
    setFailed([]);
    setError("");
    setNotice("");
    const failures: { file: File; error: string }[] = [];
    try {
      for (const [i, file] of selected.entries()) {
        setUploading(`Uploading ${i + 1} of ${selected.length}: ${file.name}`);
        try {
          const data = new FormData();
          data.set("paper_id", paperID);
          data.set("template_id", template);
          data.set("file", file);
          await api.uploadOMRCheck(data);
          setRevision((n) => n + 1);
        } catch (e) {
          failures.push({
            file,
            error: e instanceof Error ? e.message : "Upload failed",
          });
        }
      }
      setFailed(failures);
      setNotice(
        `${selected.length - failures.length} file(s) accepted. Existing uploads are reused; pending sheets are read and graded in the background.`,
      );
    } finally {
      setBusy(false);
      setUploading("");
    }
  }
  async function saveTemplate(profile: OMRProfile) {
    if (!paperID) throw Error("Select an assessment before saving a template.");
    setSaving(true);
    try {
      const data = new FormData();
      data.set("paper_id", paperID);
      data.set("file", referenceFile(profile));
      data.set(
        "definition",
        JSON.stringify({
          name: profile.layout.name,
          template: profile.template,
          config: profile.config,
          layout: profile.layout,
          identity: Object.entries(profile.layout.identity).map(
            ([kind, rect]) => ({ kind, rect }),
          ),
        }),
      );
      const saved = await api.saveOMRTemplate(data);
      setTemplates((old) => [...old, saved]);
      selectTemplate(saved.id);
      setNotice(
        `“${saved.name}” is saved and selected. Upload answer sheets above.`,
      );
    } finally {
      setSaving(false);
    }
  }
  const compatible = templates.filter(
    (t) => paper && t.question_count === questions(paper).length,
  );
  return (
    <>
      <section
        className="card omrchecker-reader"
        aria-label="OMR sheet scanner"
      >
        <h2>Check answer sheets</h2>
        <p className="muted">
          Choose the assessment and saved template, then upload one sheet or a
          batch. Students do not need an account.
        </p>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void upload(files);
          }}
        >
          <fieldset disabled={busy || saving} className="omrchecker-inputs">
            <div className="academic-columns">
              <Field name="1. Assessment">
                <select
                  value={paperID}
                  required
                  onChange={(e) =>
                    onPaper(papers.find((p) => p.id === e.target.value) ?? null)
                  }
                >
                  <option value="">Choose an assessment</option>
                  {papers
                    .filter((p) => p.state === "published")
                    .map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.spec.title} · {p.reference}
                      </option>
                    ))}
                </select>
              </Field>
              <Field name="2. OMR template">
                <select
                  value={template}
                  required
                  disabled={!paperID || loading}
                  onChange={(e) => selectTemplate(e.target.value)}
                >
                  <option value="">
                    {loading ? "Loading templates…" : "Choose a saved template"}
                  </option>
                  {compatible.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} · {t.question_count} questions
                    </option>
                  ))}
                </select>
              </Field>
            </div>
            {paperID && !loading && !compatible.length && (
              <p className="muted">
                No matching templates yet. Create and name one below; it will be
                available here next time.
              </p>
            )}
            <Field name="3. Answer sheets">
              <input
                key={paperID}
                type="file"
                accept="image/png,image/jpeg,application/pdf"
                multiple
                required
                disabled={!paperID || !template}
                onChange={(e) => {
                  setFiles(Array.from(e.target.files ?? []));
                  setFailed([]);
                  setError("");
                }}
              />
            </Field>
            <p className="muted">
              Select a single file or several together. PNG/JPEG: up to 8 MiB
              each. PDF: up to 50 MiB and 20 pages. Up to 20 files per upload.
            </p>
            <button
              className="primary"
              disabled={!paperID || !template || !files.length || loading}
            >
              {busy ? "Uploading sheets…" : "Check answer sheets"}
            </button>
          </fieldset>
        </form>
        {uploading && <p role="status">{uploading}</p>}
        {notice && (
          <p role="status" className="omr-flow-notice">
            {notice}
          </p>
        )}
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        {!!failed.length && (
          <div role="alert">
            <p>These files were not uploaded. The other files are retained.</p>
            <ul>
              {failed.map((f, i) => (
                <li key={i}>
                  {f.file.name}: {f.error}
                </li>
              ))}
            </ul>
            <button
              type="button"
              disabled={busy}
              onClick={() => void upload(failed.map((f) => f.file))}
            >
              Retry failed uploads
            </button>
          </div>
        )}
      </section>
      {paperID && paper && (
        <OMRResults key={paperID} api={api} paper={paper} revision={revision} />
      )}
      <details className="card omr-template-creation">
        <summary>Create an OMR template</summary>
        {!paperID ? (
          <p>Select an assessment above to create a matching template.</p>
        ) : (
          <>
            <p className="muted">
              Upload a blank or sample sheet, draw and label its fields, then
              save it with a name. The template is reusable for this batch.
            </p>
            <Field name="Template sheet image">
              <input
                type="file"
                accept="image/png,image/jpeg"
                disabled={saving || busy}
                onChange={(e) => setSample(e.target.files?.[0] ?? null)}
              />
            </Field>
            <OMRTemplateBuilder
              key={paperID}
              file={sample}
              onProfile={() => {}}
              onSave={saveTemplate}
            />
          </>
        )}
      </details>
    </>
  );
}

export function OMRResults({
  api,
  paper,
  revision = 0,
}: {
  api: EducationClient;
  paper: Paper;
  revision?: number;
}) {
  const [checks, setChecks] = useState<Check[]>([]),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(true);
  const [refresh, setRefresh] = useState(0),
    [review, setReview] = useState<Sheet | null>(null),
    [opening, setOpening] = useState(false),
    [query, setQuery] = useState("");
  const [cursors, setCursors] = useState([""]);
  const cursor = cursors.at(-1) ?? "";
  useEffect(() => setCursors([""]), [revision]);
  useEffect(() => {
    let active = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    async function load() {
      try {
        const result = await api.list("omr-checks", {
          paper: paper.id,
          before: cursor,
        });
        if (!active) return;
        setChecks(result);
        setError("");
        if (result.some((c) => ["queued", "running"].includes(c.state)))
          timer = setTimeout(() => void load(), 2000);
      } catch (e) {
        if (active)
          setError(
            e instanceof Error ? e.message : "Could not load sheet results.",
          );
      } finally {
        if (active) setLoading(false);
      }
    }
    void load();
    return () => {
      active = false;
      if (timer) clearTimeout(timer);
    };
  }, [api, paper.id, revision, refresh, cursor]);
  async function open(id: string) {
    setOpening(true);
    try {
      setReview(await api.detail("omr-sheets", id));
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not open this sheet.");
    } finally {
      setOpening(false);
    }
  }
  async function control(id: string, action: "retry" | "cancel") {
    try {
      await api.jobAction(id, action);
      setRefresh((n) => n + 1);
    } catch (e) {
      setError(
        e instanceof Error ? e.message : "Could not update this upload.",
      );
    }
  }
  const lower = query.trim().toLowerCase();
  const count = checks
    .flatMap((c) => c.sheets)
    .filter((s) => s.state === "graded").length;
  return (
    <section
      className="card omr-saved-results"
      aria-label="Registration-based OMR results"
      aria-busy={loading}
    >
      <div className="section-heading">
        <div>
          <h2>Sheet results</h2>
          <p className="muted">
            {count} graded · Registration number and name come from the sheet.
            No account matching is required.
          </p>
        </div>
        <button type="button" onClick={() => setRefresh((n) => n + 1)}>
          Refresh sheet results
        </button>
      </div>
      <Field name="Search registration or name">
        <input
          type="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Registration number or student name"
        />
      </Field>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {loading && <p role="status">Loading saved sheet results…</p>}
      {!loading && !checks.length && !error && (
        <p className="empty-message">
          Uploaded sheets and their grades will appear here.
        </p>
      )}
      {!!checks.length && (
        <div
          className="results-table-wrap"
          tabIndex={0}
          role="region"
          aria-label="OMR grades by registration number"
        >
          <table className="results-table">
            <thead>
              <tr>
                <th scope="col">Registration number</th>
                <th scope="col">Student name</th>
                <th scope="col">Marks</th>
                <th scope="col">Correct questions</th>
                <th scope="col">Wrong questions</th>
                <th scope="col">Status</th>
                <th scope="col">Review</th>
              </tr>
            </thead>
            <tbody>
              {checks.flatMap((c) =>
                c.sheets
                  .filter(
                    (s) =>
                      !lower ||
                      `${s.registration_number} ${s.student_name} ${c.filename}`
                        .toLowerCase()
                        .includes(lower),
                  )
                  .map((s) => (
                    <tr key={s.id}>
                      <th scope="row">
                        {s.registration_number || "Not detected"}
                        <small className="muted omr-result-source">
                          {c.filename} · page {s.page}
                        </small>
                      </th>
                      <td>{s.student_name || "—"}</td>
                      <td>
                        {s.state === "excluded"
                          ? "—"
                          : `${s.score} / ${s.maximum}`}
                        {s.state === "needs_review" && (
                          <small className="omr-result-source">
                            Provisional
                          </small>
                        )}
                      </td>
                      <td>{s.correct_questions.join(", ") || "—"}</td>
                      <td>{s.wrong_questions.join(", ") || "—"}</td>
                      <td>
                        {s.state === "graded"
                          ? "Graded"
                          : s.state === "excluded"
                            ? "Excluded"
                            : "Needs review"}
                        {s.identity_uncertain && (
                          <small className="omr-result-source">
                            Check identity
                          </small>
                        )}
                      </td>
                      <td>
                        <button
                          type="button"
                          disabled={opening}
                          onClick={() => void open(s.id)}
                        >
                          {s.state === "needs_review"
                            ? "Review sheet"
                            : "View sheet"}
                        </button>
                      </td>
                    </tr>
                  )),
              )}
            </tbody>
          </table>
        </div>
      )}
      {(checks.length === 100 || cursors.length > 1) && (
        <div className="actions">
          <span className="muted">
            Up to 100 uploads per page. Search applies to this page.
          </span>
          <button
            type="button"
            disabled={cursors.length === 1}
            onClick={() => setCursors((old) => old.slice(0, -1))}
          >
            Newer uploads
          </button>
          <button
            type="button"
            disabled={checks.length < 100}
            onClick={() => setCursors((old) => [...old, checks.at(-1)!.id])}
          >
            Older uploads
          </button>
        </div>
      )}
      {checks
        .filter((c) => c.state !== "completed")
        .map((c) => (
          <div className="omr-upload-status" key={c.id}>
            <div>
              <strong>{c.filename}</strong>
              <p>
                {c.state.replaceAll("_", " ")} · {c.progress}%
                {c.sheets.length
                  ? ` · ${c.sheets.length} page(s) retained`
                  : ""}
              </p>
              {c.error && <p className="error">{c.error}</p>}
            </div>
            {["queued", "running"].includes(c.state) && (
              <button
                type="button"
                onClick={() => void control(c.job_id, "cancel")}
              >
                Cancel upload
              </button>
            )}
            {["failed", "configuration_required", "cancelled"].includes(
              c.state,
            ) && (
              <button
                type="button"
                onClick={() => void control(c.job_id, "retry")}
              >
                Retry reading
              </button>
            )}
          </div>
        ))}
      {review && (
        <OMRSheetReview
          key={review.summary.id + ":" + review.summary.version}
          api={api}
          paper={paper}
          sheet={review}
          onClose={() => setReview(null)}
          onSaved={() => {
            setReview(null);
            setRefresh((n) => n + 1);
          }}
        />
      )}
    </section>
  );
}
function OMRSheetReview({
  api,
  paper,
  sheet,
  onClose,
  onSaved,
}: {
  api: EducationClient;
  paper: Paper;
  sheet: Sheet;
  onClose: () => void;
  onSaved: () => void;
}) {
  const [registration, setRegistration] = useState(
      sheet.summary.registration_number,
    ),
    [name, setName] = useState(sheet.summary.student_name);
  const [answers, setAnswers] = useState(sheet.answers),
    [confirmed, setConfirmed] = useState(false),
    [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function save(exclude = false) {
    setBusy(true);
    setError("");
    try {
      await api.reviewOMRSheet(sheet.summary.id, {
        version: sheet.summary.version,
        registration_number: registration,
        student_name: name,
        confirmed,
        reason,
        exclude,
        answers,
      });
      onSaved();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save this review.");
    } finally {
      setBusy(false);
    }
  }
  return (
    <section className="omr-sheet-review" aria-label="Review scanned sheet">
      <div className="section-heading">
        <h3>Review page {sheet.summary.page}</h3>
        <button type="button" disabled={busy} onClick={onClose}>
          Close sheet review
        </button>
      </div>
      <p className="muted">
        Check the detected identity and highlighted responses against the sheet.
        Registration numbers keep leading zeros.
      </p>
      <div className="academic-columns">
        <div>
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void save();
            }}
          >
            <fieldset disabled={busy} className="omrchecker-inputs">
              <Field name="Registration number">
                <input
                  value={registration}
                  maxLength={100}
                  onChange={(e) => setRegistration(e.target.value)}
                />
              </Field>
              <Field name="Student name">
                <input
                  value={name}
                  maxLength={160}
                  onChange={(e) => setName(e.target.value)}
                />
              </Field>
              <details open={sheet.summary.state === "needs_review"}>
                <summary>
                  Answers · {sheet.summary.uncertain_questions.length} need
                  review
                </summary>
                {questions(paper).map((q, i) => {
                  const a = answers[q.id!] ?? {
                    selection: [],
                    number: "",
                    text: "",
                    visited: true,
                    review: false,
                    order: i + 1,
                    invalid: false,
                  };
                  return (
                    <div
                      className={
                        sheet.summary.uncertain_questions.includes(i + 1)
                          ? "omr-answer-review flagged"
                          : "omr-answer-review"
                      }
                      key={q.id}
                    >
                      <strong>
                        Question {i + 1}
                        {sheet.summary.uncertain_questions.includes(i + 1)
                          ? " · check reading"
                          : ""}
                      </strong>
                      {q.type === "single" ? (
                        <Field name={`Answer ${i + 1}`}>
                          <select
                            value={
                              a.invalid ? "invalid" : (a.selection[0] ?? "")
                            }
                            onChange={(e) =>
                              setAnswers({
                                ...answers,
                                [q.id!]: {
                                  ...a,
                                  selection:
                                    e.target.value &&
                                    e.target.value !== "invalid"
                                      ? [e.target.value]
                                      : [],
                                  invalid: e.target.value === "invalid",
                                  review: false,
                                },
                              })
                            }
                          >
                            <option value="">Blank</option>
                            {q.options.map((o, j) => (
                              <option key={o.id} value={o.id}>
                                {j + 1} · {o.id}
                              </option>
                            ))}
                            <option value="invalid">
                              Invalid / multiple marks
                            </option>
                          </select>
                        </Field>
                      ) : (
                        <fieldset>
                          <legend>Answer {i + 1}</legend>
                          {q.options.map((o, j) => (
                            <label className="check" key={o.id}>
                              <input
                                type="checkbox"
                                checked={a.selection.includes(o.id)}
                                onChange={(e) =>
                                  setAnswers({
                                    ...answers,
                                    [q.id!]: {
                                      ...a,
                                      selection: e.target.checked
                                        ? [...a.selection, o.id]
                                        : a.selection.filter((v) => v !== o.id),
                                      invalid: false,
                                      review: false,
                                    },
                                  })
                                }
                              />
                              {j + 1} · {o.id}
                            </label>
                          ))}
                          <label className="check">
                            <input
                              type="checkbox"
                              checked={a.invalid}
                              onChange={(e) =>
                                setAnswers({
                                  ...answers,
                                  [q.id!]: {
                                    ...a,
                                    invalid: e.target.checked,
                                    review: false,
                                  },
                                })
                              }
                            />
                            Invalid response
                          </label>
                        </fieldset>
                      )}
                    </div>
                  );
                })}
              </details>
              <Field name="Sheet review reason">
                <textarea
                  required
                  maxLength={2000}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                />
              </Field>
              <label className="check">
                <input
                  type="checkbox"
                  required
                  checked={confirmed}
                  onChange={(e) => setConfirmed(e.target.checked)}
                />
                I checked the identity and responses against this sheet.
              </label>
              {error && (
                <p role="alert" className="error">
                  {error}
                </p>
              )}
              <div className="actions">
                <button className="primary">
                  {busy ? "Saving…" : "Save reviewed grade"}
                </button>
                <button type="button" onClick={() => void save(true)}>
                  Exclude this sheet
                </button>
              </div>
            </fieldset>
          </form>
        </div>
        <div>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={api.url("assets/" + sheet.asset_id)}
            alt="Scanned sheet with detected bubbles marked"
          />
          <p className="muted">
            Read locally by {sheet.reading.engine} ·{" "}
            {sheet.reading.revision.slice(0, 7)}
          </p>
        </div>
      </div>
    </section>
  );
}
