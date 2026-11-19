"use client";
import { useEffect, useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { selectQuestionNumbers } from "./question-selection";
import { TaskProgress } from "./task-progress";
import { FormField as Field } from "./form-field";
type Import = A<"AcademicImport">;
type Job = A<"AcademicJob">;

export function ImportChoiceRow({
  row,
  disabled,
  onSave,
  onAdvanced,
}: {
  row: Import["rows"][number];
  disabled: boolean;
  onSave: (q: A<"AcademicQuestion">) => void;
  onAdvanced: () => void;
}) {
  const [q, setQ] = useState(row.question);
  return (
    <article className="import-question">
      <div className="section-heading">
        <h3>Question {row.source_number || row.index}</h3>
        <span className="state-pill">
          {row.committed_id ? "Saved to bank" : "Faculty review"}
        </span>
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          onSave(q);
        }}
      >
        <fieldset disabled={disabled || !!row.committed_id}>
          <legend className="sr-only">
            Review question {row.source_number || row.index}
          </legend>
          <Field name={`Question ${row.index} wording`}>
            <textarea
              value={q.stem}
              required
              onChange={(e) => setQ({ ...q, stem: e.target.value })}
            />
          </Field>
          <p className="muted">
            Edit the extracted options. Tick the correct answer after checking
            the source, or apply a separate key below.
          </p>
          {q.options?.map((o, i) => (
            <div className="import-option" key={o.id}>
              <label className="check">
                <input
                  type={q.type === "single" ? "radio" : "checkbox"}
                  name={`import-key-${row.index}`}
                  checked={
                    q.type === "single" && (q.key?.length ?? 0) > 1
                      ? false
                      : (q.key?.includes(o.id) ?? false)
                  }
                  onChange={(e) =>
                    setQ({
                      ...q,
                      key:
                        q.type === "single"
                          ? [o.id]
                          : e.target.checked
                            ? [...(q.key ?? []), o.id]
                            : (q.key ?? []).filter((k) => k !== o.id),
                    })
                  }
                />
                Correct {o.id}
              </label>
              <Field name={`Question ${row.index} option ${o.id}`}>
                <input
                  required
                  value={o.text}
                  onChange={(e) =>
                    setQ({
                      ...q,
                      options: q.options!.map((old, j) =>
                        j === i ? { ...old, text: e.target.value } : old,
                      ),
                    })
                  }
                />
              </Field>
            </div>
          ))}
          <details>
            <summary>Marks and explanation</summary>
            <div className="academic-columns">
              <Field name={`Question ${row.index} marks`}>
                <input
                  value={q.marks}
                  required
                  onChange={(e) => setQ({ ...q, marks: e.target.value })}
                />
              </Field>
              <Field name={`Question ${row.index} penalty`}>
                <input
                  value={q.penalty}
                  required
                  onChange={(e) => setQ({ ...q, penalty: e.target.value })}
                />
              </Field>
            </div>
            <Field name={`Question ${row.index} explanation`}>
              <textarea
                value={q.solution}
                onChange={(e) => setQ({ ...q, solution: e.target.value })}
              />
            </Field>
          </details>
          {row.error && (
            <p className="muted">
              {row.error === "supply the approved selection key"
                ? "Choose the correct answer, then save this question."
                : row.error}
            </p>
          )}
          <div className="actions">
            <button>Save reviewed question {row.index}</button>
            <button type="button" onClick={onAdvanced}>
              Correct row {row.index}
            </button>
          </div>
        </fieldset>
      </form>
    </article>
  );
}

export function KeyImportPanel({
  preview,
  api,
  jobs,
  onApplied,
  simple = false,
}: {
  simple?: boolean;
  preview: Import;
  api: EducationClient;
  jobs: Job[];
  onApplied: (p: Import) => void;
}) {
  const [job, setJob] = useState<Job | null>(
    () =>
      jobs.find(
        (j) =>
          j.resource_id === preview.id &&
          (j.input as Record<string, unknown>)?.mode === "answer_key",
      ) ?? null,
  );
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  useEffect(() => {
    if (!job || !["queued", "running"].includes(job.state)) return;
    let active = true,
      pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        const next = await api.detail("jobs", job.id);
        if (active) {
          setJob(next);
          setError("");
        }
      } catch (e) {
        if (active)
          setError(
            e instanceof Error
              ? e.message
              : "Check the connection, then refresh status.",
          );
      } finally {
        pending = false;
      }
    }, 2000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, job?.id, job?.state]);
  const output = job?.output as A<"AcademicKeyPreview"> | undefined;
  return (
    <section className="key-import-panel">
      <h3>Have a separate answer key?</h3>
      <p>
        Upload CSV, text or JSON, or a clearly printed PDF/image. Use one entry
        per line: <code>1,B</code>. Numeric labels 1–8 map to A–H. Review the
        extracted key before applying it; mismatched question numbers are
        rejected.
      </p>
      {(busy || (job && ["queued", "running"].includes(job.state))) && (
        <TaskProgress
          state={busy ? "uploading" : job!.state}
          title={
            busy
              ? "Saving your answer key"
              : "Extracting the separate answer key"
          }
          detail="Questions are retained. Check the numbered key preview before applying it."
          progress={busy ? undefined : job?.progress}
        />
      )}
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          const form = new FormData(e.currentTarget);
          setBusy(true);
          setError("");
          try {
            setJob(await api.uploadKey(preview.id, form));
          } catch (e) {
            setError(
              e instanceof Error
                ? e.message
                : "Key upload failed. Your questions are retained.",
            );
          } finally {
            setBusy(false);
          }
        }}
      >
        {simple ? (
          <input type="hidden" name="engine" value="tesseract" />
        ) : (
          <>
            <Field name="Answer key extraction engine">
              <select name="engine">
                <option value="">Default engine</option>
                <option value="lightonocr">
                  LightOnOCR-2-1B (local model)
                </option>
                <option value="tesseract">Tesseract (local text OCR)</option>
              </select>
            </Field>
          </>
        )}
        <Field name="Separate answer key file">
          <input
            type="file"
            name="file"
            accept=".csv,.txt,.json,.pdf,.png,.jpg,.jpeg"
            required
          />
        </Field>
        <button
          disabled={
            busy ||
            preview.state !== "preview" ||
            (!!job && ["queued", "running"].includes(job.state))
          }
        >
          {busy
            ? "Saving key…"
            : job && ["queued", "running"].includes(job.state)
              ? "Reading key…"
              : "Read separate answer key"}
        </button>
      </form>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {job && (
        <div aria-live="polite">
          <p>
            Answer key: {job.state}
            {job.error ? ` — ${job.error}` : ""}
          </p>
          {job.state === "completed" && Array.isArray(output?.entries) && (
            <>
              <div className="key-entry-list">
                {output.entries.map((e) => (
                  <span key={e.number}>
                    {e.number}: {e.selection.join(" + ")}
                  </span>
                ))}
              </div>
              <p>
                Applying this key updates the saved preview. Save any unsaved
                question edits first.
              </p>
              <button
                disabled={busy || preview.state !== "preview"}
                onClick={async () => {
                  setBusy(true);
                  setError("");
                  try {
                    onApplied(
                      await api.applyKey(preview.id, {
                        job_id: job.id,
                        version: preview.version,
                      }),
                    );
                  } catch (e) {
                    setError(
                      e instanceof Error ? e.message : "Could not apply key.",
                    );
                  } finally {
                    setBusy(false);
                  }
                }}
              >
                I reviewed this key — apply to questions
              </button>
            </>
          )}
          {["failed", "cancelled", "configuration_required"].includes(
            job.state,
          ) && (
            <button
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                try {
                  await api.jobAction(job.id, "retry");
                  setJob(await api.detail("jobs", job.id));
                } catch (e) {
                  setError(e instanceof Error ? e.message : "Retry failed");
                } finally {
                  setBusy(false);
                }
              }}
            >
              Retry key extraction
            </button>
          )}
        </div>
      )}
    </section>
  );
}

export function MarkingRules({
  numbers,
  disabled,
  onApply,
}: {
  numbers: number[];
  disabled: boolean;
  onApply: (numbers: number[], marks: string, penalty: string) => Promise<void>;
}) {
  const [selection, setSelection] = useState("");
  const [all, setAll] = useState(false);
  const [marks, setMarks] = useState("4");
  const [penalty, setPenalty] = useState("1");
  const [error, setError] = useState("");
  let selected: number[] = [];
  try {
    if (all || selection.trim())
      selected = selectQuestionNumbers(
        all ? numbers.join(",") : selection,
        numbers,
      );
  } catch {
    /* Validation appears on submit. */
  }
  return (
    <section className="marking-rules" aria-label="Marking scheme">
      <div className="section-heading">
        <div>
          <h3>Set the marking scheme</h3>
          <p className="muted">
            Apply marks to a range or specific questions. Blank answers receive
            0.
          </p>
        </div>
        <span className="state-pill">{numbers.length} questions</span>
      </div>
      <label className="check">
        <input
          type="checkbox"
          checked={all}
          onChange={(e) => setAll(e.target.checked)}
        />
        All questions
      </label>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          setError("");
          try {
            await onApply(
              selectQuestionNumbers(
                all ? numbers.join(",") : selection,
                numbers,
              ),
              marks,
              penalty,
            );
          } catch (e) {
            setError(
              e instanceof Error ? e.message : "Could not save marking rules.",
            );
          }
        }}
      >
        <fieldset disabled={disabled} className="marking-fields">
          <Field name="Question numbers">
            <input
              value={selection}
              onChange={(e) => setSelection(e.target.value)}
              placeholder="1-10 or 1,5,8,14,25"
              disabled={all}
              required={!all}
              aria-describedby="range-help"
            />
          </Field>
          <Field name="Marks for correct answer">
            <input
              value={marks}
              onChange={(e) => setMarks(e.target.value)}
              inputMode="decimal"
              pattern="[0-9]+([.][0-9]+)?"
              required
            />
          </Field>
          <Field name="Deduction for wrong answer">
            <input
              value={penalty}
              onChange={(e) => setPenalty(e.target.value)}
              inputMode="decimal"
              pattern="[0-9]+([.][0-9]+)?"
              required
            />
          </Field>
          <button type="submit">Apply marks</button>
        </fieldset>
        <small id="range-help" className="muted">
          {selected.length
            ? `${selected.length} selected: ${selected.join(", ")} · +${marks} correct / −${penalty} wrong`
            : "Use the question numbers shown below. Save any question edits before applying marks."}
        </small>
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
      </form>
    </section>
  );
}
