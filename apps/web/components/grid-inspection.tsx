"use client";
import { useEffect, useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { TaskProgress } from "./task-progress";
import { FormField as Field } from "./form-field";
type Calibration = A<"AcademicGridCalibration">;
const corners = [
  "top_left",
  "top_right",
  "bottom_left",
  "bottom_right",
] as const;
const cornerNames = [
  "first row, first bubble",
  "first row, last bubble",
  "last row, first bubble",
  "last row, last bubble",
];
function initialCalibration(): Calibration {
  return {
    options: 4,
    radius: 0.008,
    blocks: Array.from({ length: 4 }, (_, i) => {
      const x = (63 + i * 157.2) / 638;
      return {
        first: i * 45 + 1,
        count: 45,
        top_left: { x, y: 37 / 1208 },
        top_right: { x: x + 79 / 638, y: 37 / 1208 },
        bottom_left: { x: x + 5 / 638, y: 1188 / 1208 },
        bottom_right: { x: x + 84 / 638, y: 1188 / 1208 },
      };
    }),
  };
}
export function GridInspection({
  api,
  courses,
}: {
  api: EducationClient;
  courses: A<"AcademicCourse">[];
}) {
  const [file, setFile] = useState<File | null>(null);
  const [url, setUrl] = useState("");
  const [calibration, setCalibration] = useState(initialCalibration);
  const [target, setTarget] = useState<{
    block: number;
    corner: number;
  } | null>(null);
  const [reading, setReading] = useState<A<"AcademicGridReading"> | null>(null);
  const [review, setReview] = useState<Record<number, string>>({});
  const [busy, setBusy] = useState(false);
  const [modelCheck, setModelCheck] = useState(false);
  const [modelJob, setModelJob] = useState<A<"AcademicJob"> | null>(null);
  const [error, setError] = useState("");
  useEffect(() => {
    if (!file) {
      setUrl("");
      return;
    }
    const next = URL.createObjectURL(file);
    setUrl(next);
    return () => URL.revokeObjectURL(next);
  }, [file]);
  useEffect(() => {
    if (!modelJob || !["queued", "running"].includes(modelJob.state)) return;
    let active = true,
      pending = false;
    const timer = setInterval(async () => {
      if (pending) return;
      pending = true;
      try {
        const next = await api.detail("jobs", modelJob.id);
        if (active) {
          setModelJob(next);
          if (next.state === "completed") {
            const out = next.output as A<"AcademicGridModelOutput">;
            if (Array.isArray(out?.reading?.rows)) setReading(out.reading);
          }
        }
      } catch (e) {
        if (active)
          setError(
            e instanceof Error
              ? e.message
              : "Could not refresh model progress.",
          );
      } finally {
        pending = false;
      }
    }, 2000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, modelJob?.id, modelJob?.state]);
  function change(c: Calibration) {
    setCalibration(c);
    setModelJob(null);
    setReading(null);
    setReview({});
  }
  return (
    <details className="card grid-inspection">
      <summary>Pixel-reader diagnostics</summary>
      <p>
        For sheets without our registration squares, align the bubble centers
        before reading. This preview reads answers only: it cannot identify a
        learner or award marks. Use the visual template editor above to create
        OMRChecker layouts. These older diagnostics accept explicit grid
        coordinates.
      </p>
      <form
        onSubmit={async (e) => {
          e.preventDefault();
          if (!file) return;
          setBusy(true);
          setError("");
          try {
            const f = new FormData(e.currentTarget);
            f.set("file", file);
            f.set("calibration", JSON.stringify(calibration));
            setReading(await api.previewGrid(f));
            setReview({});
            if (modelCheck) {
              const next = await api.modelGrid(f);
              setModelJob(next);
              if (next.state === "completed") {
                const out = next.output as A<"AcademicGridModelOutput">;
                if (Array.isArray(out?.reading?.rows)) setReading(out.reading);
              }
            }
          } catch (e) {
            setError(
              e instanceof Error ? e.message : "Could not read the image.",
            );
          } finally {
            setBusy(false);
          }
        }}
      >
        <div className="academic-columns">
          <Field name="OMR preview course">
            <select name="course_id" required>
              <option value="">Choose course</option>
              {courses.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </Field>
          <Field name="Existing OMR image">
            <input
              type="file"
              accept="image/png,image/jpeg"
              required
              onChange={(e) => {
                setFile(e.target.files?.[0] ?? null);
                setModelJob(null);
                setReading(null);
                setReview({});
              }}
            />
          </Field>
        </div>
        <div className="academic-columns">
          <Field name="Options in each row">
            <input
              type="number"
              min={2}
              max={8}
              value={calibration.options}
              onChange={(e) =>
                change({ ...calibration, options: Number(e.target.value) })
              }
              required
            />
          </Field>
          <Field name="Sampling radius (% of image width)">
            <input
              type="number"
              min={0.1}
              max={3}
              step={0.1}
              value={Number((calibration.radius * 100).toFixed(2))}
              onChange={(e) =>
                change({ ...calibration, radius: Number(e.target.value) / 100 })
              }
              required
            />
          </Field>
        </div>
        {url && (
          <div>
            {target && (
              <p role="status">
                Click the center of the {cornerNames[target.corner]} in block{" "}
                {target.block + 1}, or enter coordinates below.{" "}
                <button type="button" onClick={() => setTarget(null)}>
                  Cancel point
                </button>
              </p>
            )}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={url}
              alt="Source for legacy grid diagnostics"
              onClick={(e) => {
                const box = e.currentTarget.getBoundingClientRect();
                const p = {
                  x: (e.clientX - box.left) / box.width,
                  y: (e.clientY - box.top) / box.height,
                };
                if (!target) return;
                change({
                  ...calibration,
                  blocks: calibration.blocks.map((b, i) =>
                    i === target.block
                      ? { ...b, [corners[target.corner]]: p }
                      : b,
                  ),
                });
                setTarget(null);
              }}
            />
          </div>
        )}
        <details className="omr-advanced">
          <summary>Advanced grid coordinates</summary>
          <div className="grid-calibration-columns">
            <div>
              {calibration.blocks.map((b, index) => (
                <details key={index}>
                  <summary>
                    Block {index + 1}: questions {b.first}–
                    {b.first + b.count - 1}
                  </summary>
                  <div className="academic-columns">
                    <Field name={`Block ${index + 1} first question`}>
                      <input
                        type="number"
                        min={1}
                        max={500}
                        value={b.first}
                        onChange={(e) =>
                          change({
                            ...calibration,
                            blocks: calibration.blocks.map((old, i) =>
                              i === index
                                ? { ...old, first: Number(e.target.value) }
                                : old,
                            ),
                          })
                        }
                      />
                    </Field>
                    <Field name={`Block ${index + 1} rows`}>
                      <input
                        type="number"
                        min={1}
                        max={200}
                        value={b.count}
                        onChange={(e) =>
                          change({
                            ...calibration,
                            blocks: calibration.blocks.map((old, i) =>
                              i === index
                                ? { ...old, count: Number(e.target.value) }
                                : old,
                            ),
                          })
                        }
                      />
                    </Field>
                  </div>
                  {corners.map((corner, ci) => (
                    <div className="calibration-point" key={corner}>
                      <button
                        type="button"
                        aria-pressed={
                          target?.block === index && target.corner === ci
                        }
                        onClick={() => setTarget({ block: index, corner: ci })}
                      >
                        Mark {cornerNames[ci]}
                      </button>
                      <div className="academic-columns">
                        {(["x", "y"] as const).map((axis) => (
                          <Field
                            key={axis}
                            name={`Block ${index + 1} ${corner} ${axis} percent`}
                          >
                            <input
                              type="number"
                              min={0}
                              max={100}
                              step={0.001}
                              value={Number((b[corner][axis] * 100).toFixed(3))}
                              onChange={(e) =>
                                change({
                                  ...calibration,
                                  blocks: calibration.blocks.map((old, i) =>
                                    i === index
                                      ? {
                                          ...old,
                                          [corner]: {
                                            ...old[corner],
                                            [axis]:
                                              Number(e.target.value) / 100,
                                          },
                                        }
                                      : old,
                                  ),
                                })
                              }
                            />
                          </Field>
                        ))}
                      </div>
                    </div>
                  ))}
                  <button
                    type="button"
                    disabled={calibration.blocks.length === 1}
                    onClick={() =>
                      change({
                        ...calibration,
                        blocks: calibration.blocks.filter(
                          (_, i) => i !== index,
                        ),
                      })
                    }
                  >
                    Remove block {index + 1}
                  </button>
                </details>
              ))}
              <button
                type="button"
                disabled={calibration.blocks.length >= 8}
                onClick={() =>
                  change({
                    ...calibration,
                    blocks: [
                      ...calibration.blocks,
                      {
                        ...initialCalibration().blocks[0],
                        first: Math.max(
                          ...calibration.blocks.map((b) => b.first + b.count),
                        ),
                        count: 45,
                      },
                    ],
                  })
                }
              >
                Add response block
              </button>
            </div>
          </div>
        </details>
        <label className="check">
          <input
            type="checkbox"
            checked={modelCheck}
            onChange={(e) => setModelCheck(e.target.checked)}
          />
          Also compare with a local mark model
        </label>
        <p className="muted">
          The default classifier was trained on synthetic marks; field
          qualification is still pending. Model readings are proposals.
          Disagreements with measured fill require review. The image is retained
          privately for a durable model job; the pixel-only diagnostic is not
          stored.
        </p>
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        <button className="primary" disabled={!file || busy}>
          {busy ? "Reading bubbles…" : "Read calibrated answers"}
        </button>
      </form>
      {(busy ||
        (modelJob && ["queued", "running"].includes(modelJob.state))) && (
        <TaskProgress
          state={busy ? "uploading" : modelJob!.state}
          title={
            busy
              ? "Reading calibrated bubbles"
              : "Comparing detected marks with the model"
          }
          detail="Measured responses remain available. Confirm blanks, uncertain marks and any model disagreements before using the reading."
          progress={busy ? undefined : modelJob?.progress}
        />
      )}
      {modelJob && (
        <section className="model-inspection-state" aria-live="polite">
          <h3>
            Vision comparison:{" "}
            {{
              queued: "Waiting to start",
              running: "Comparing responses",
              completed: "Ready for review",
              failed: "Could not finish",
              configuration_required: "The vision service needs setup",
              cancelled: "Cancelled",
            }[modelJob.state] ?? modelJob.state}
          </h3>
          {modelJob.model && <p className="muted">Model: {modelJob.model}</p>}
          <p>
            {modelJob.error ||
              "The saved comparison runs in the background. Pixel readings remain available below."}
          </p>
          {["queued", "running"].includes(modelJob.state) && (
            <button
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                try {
                  await api.jobAction(modelJob.id, "cancel");
                  setModelJob(await api.detail("jobs", modelJob.id));
                } catch (e) {
                  setError(
                    e instanceof Error
                      ? e.message
                      : "Could not cancel the comparison.",
                  );
                } finally {
                  setBusy(false);
                }
              }}
            >
              Cancel model comparison
            </button>
          )}
          {["failed", "configuration_required", "cancelled"].includes(
            modelJob.state,
          ) && (
            <button
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                try {
                  await api.jobAction(modelJob.id, "retry");
                  setModelJob(await api.detail("jobs", modelJob.id));
                } catch (e) {
                  setError(e instanceof Error ? e.message : "Retry failed");
                } finally {
                  setBusy(false);
                }
              }}
            >
              Retry model comparison
            </button>
          )}
        </section>
      )}
      {reading && (
        <section aria-live="polite">
          <h3>
            {reading.rows.length} responses read ·{" "}
            {reading.rows.filter((r) => r.uncertain).length} need confirmation
          </h3>
          <p>{reading.notice}</p>
          <p>
            Preview corrections stay in this screen. Download the report to
            retain them; no assessment result is created.
          </p>
          <div className="grid-answer-list">
            {reading.rows.map((r) => (
              <div
                className={
                  r.uncertain ? "grid-answer uncertain" : "grid-answer"
                }
                key={r.number}
              >
                <Field name={`Detected answer ${r.number}`}>
                  <select
                    value={
                      review[r.number] ??
                      (r.selection.length ? r.selection.join("+") : "blank")
                    }
                    onChange={(e) =>
                      setReview({ ...review, [r.number]: e.target.value })
                    }
                  >
                    {r.selection.length > 1 && (
                      <option value={r.selection.join("+")}>
                        {r.selection.join(" + ")} — multiple
                      </option>
                    )}
                    <option value="blank">Blank</option>
                    {Array.from({ length: calibration.options }, (_, i) => (
                      <option key={i + 1} value={i + 1}>
                        {i + 1}
                      </option>
                    ))}
                  </select>
                </Field>
                <small>
                  {r.model_state && (
                    <span>
                      Model:{" "}
                      {r.model_selection?.join(" + ") ||
                        (r.model_state === "uncertain"
                          ? "needs confirmation"
                          : "blank")}{" "}
                      · {r.model_state}
                      <br />
                    </span>
                  )}
                  {review[r.number] !== undefined
                    ? "Manually reviewed"
                    : r.uncertain
                      ? r.reason
                      : "Detected"}
                </small>
              </div>
            ))}
          </div>
          <button
            onClick={() => {
              const data = {
                calibration,
                notice: reading.notice,
                model_job_id: modelJob?.id ?? null,
                model: modelJob?.model ?? null,
                model_observations:
                  modelJob?.state === "completed"
                    ? (modelJob.output as A<"AcademicGridModelOutput">)
                        .inferences
                    : [],
                rows: reading.rows.map((r) => ({
                  ...r,
                  reviewed_selection: review[r.number] ?? null,
                })),
              };
              const href = URL.createObjectURL(
                new Blob([JSON.stringify(data, null, 2)], {
                  type: "application/json",
                }),
              );
              const a = document.createElement("a");
              a.href = href;
              a.download = "omr-preview-report.json";
              a.click();
              setTimeout(() => URL.revokeObjectURL(href), 1000);
            }}
          >
            Download reading and review report
          </button>
        </section>
      )}
    </details>
  );
}
