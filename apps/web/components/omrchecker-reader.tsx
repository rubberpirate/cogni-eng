"use client";
import { useEffect, useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { FormField as Field } from "./form-field";
import { TaskProgress } from "./task-progress";
import { OMRTemplateBuilder } from "./omr-template-builder";
import { referenceFile, type OMRProfile } from "../../../clients/omr-template";

export function OMRCheckerReader({
  api,
  courses,
}: {
  api: EducationClient;
  courses: A<"AcademicCourse">[];
}) {
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [reading, setReading] = useState<A<"AcademicOMRCheckerReading"> | null>(
    null,
  );
  const [overlay, setOverlay] = useState("");
  const [manual, setManual] = useState(false),
    [sheet, setSheet] = useState<File | null>(null);
  const [profile, setProfile] = useState<OMRProfile | null>(null);
  useEffect(() => {
    if (!reading?.overlay) {
      setOverlay("");
      return;
    }
    const bytes = Uint8Array.from(atob(reading.overlay), (c) =>
      c.charCodeAt(0),
    );
    const url = URL.createObjectURL(new Blob([bytes], { type: "image/jpeg" }));
    setOverlay(url);
    return () => URL.revokeObjectURL(url);
  }, [reading?.overlay]);
  return (
    <section className="card omrchecker-reader" aria-label="OMRChecker reader">
      <h2>Read with OMRChecker</h2>
      <p className="muted">
        Upload a sheet and create its layout visually. Saved layouts can be
        reused for the rest of your students.
      </p>
      <div className="actions omr-reader-mode">
        <button
          type="button"
          aria-pressed={!manual}
          disabled={busy}
          className={!manual ? "primary" : ""}
          onClick={() => {
            setManual(false);
            setReading(null);
          }}
        >
          Create layout visually
        </button>
        <button
          type="button"
          aria-pressed={manual}
          disabled={busy}
          onClick={() => {
            setManual(true);
            setReading(null);
          }}
        >
          Use existing files
        </button>
      </div>
      <form
        onChange={() => {
          setReading(null);
          setError("");
        }}
        onSubmit={async (e) => {
          e.preventDefault();
          setError("");
          setBusy(true);
          setReading(null);
          try {
            const data = new FormData(e.currentTarget);
            if (!manual) {
              if (!profile) throw Error("Finish the visual layout first.");
              data.set("template", JSON.stringify(profile.template));
              data.set("config", JSON.stringify(profile.config));
              data.set("reference", referenceFile(profile));
              data.set(
                "identity_regions",
                JSON.stringify(
                  Object.entries(profile.layout.identity).map(
                    ([kind, rect]) => ({ kind, rect }),
                  ),
                ),
              );
            } else {
              const template = data.get("template_file") as File;
              if (template.size > 256 * 1024)
                throw Error("Template JSON must be at most 256 KiB.");
              const text = await template.text();
              JSON.parse(text);
              data.set("template", text);
              data.delete("template_file");
              const config = data.get("config_file") as File | null;
              if (config?.size) {
                if (config.size > 64 * 1024)
                  throw Error("Tuning JSON must be at most 64 KiB.");
                const value = await config.text();
                JSON.parse(value);
                data.set("config", value);
              }
              data.delete("config_file");
              for (const name of ["marker", "reference"])
                if (!(data.get(name) as File)?.size) data.delete(name);
            }
            setReading(await api.previewOMRChecker(data));
          } catch (err) {
            setError(
              err instanceof Error ? err.message : "Could not read this sheet.",
            );
          } finally {
            setBusy(false);
          }
        }}
      >
        <fieldset disabled={busy} className="omrchecker-inputs">
          <div className="academic-columns">
            <Field name="OMRChecker student batch">
              <select name="course_id" required>
                <option value="">Choose batch</option>
                {courses.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field name="OMRChecker sheet image">
              <input
                type="file"
                name="file"
                accept="image/png,image/jpeg"
                required
                onChange={(e) => setSheet(e.target.files?.[0] ?? null)}
              />
            </Field>
            {manual && (
              <Field name="OMRChecker template JSON">
                <input
                  type="file"
                  name="template_file"
                  accept="application/json,.json"
                  required
                />
              </Field>
            )}
          </div>
          <div hidden={manual}>
            <fieldset disabled={manual} className="omrchecker-inputs">
              <OMRTemplateBuilder
                file={sheet}
                onProfile={(next) => {
                  setProfile(next);
                  setReading(null);
                  setError("");
                }}
              />
            </fieldset>
          </div>
          {manual && (
            <details>
              <summary>Optional marker, blank reference and tuning</summary>
              <div className="academic-columns">
                <Field name="OMRChecker marker image">
                  <input
                    type="file"
                    name="marker"
                    accept="image/png,image/jpeg"
                  />
                </Field>
                <Field name="OMRChecker blank reference">
                  <input
                    type="file"
                    name="reference"
                    accept="image/png,image/jpeg"
                  />
                </Field>
                <Field name="OMRChecker tuning JSON">
                  <input
                    type="file"
                    name="config_file"
                    accept="application/json,.json"
                  />
                </Field>
              </div>
            </details>
          )}
          <button className="primary" disabled={busy || (!manual && !profile)}>
            {busy ? "Reading with OMRChecker…" : "Read OMRChecker sheet"}
          </button>
        </fieldset>
      </form>
      <div className="actions omrchecker-sample-links">
        <a href="/samples/omrchecker/filled-sheet.png" download>
          Filled test sheet
        </a>
        <a href="/samples/omrchecker/template.json" download>
          Matching template
        </a>
        <a href="/samples/omrchecker/expected-answers.json" download>
          Expected filled answers
        </a>
      </div>
      {busy && (
        <TaskProgress
          state="running"
          title="OMRChecker is reading the sheet"
          detail="Applying the template's preprocessing and adaptive bubble thresholds."
        />
      )}
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {reading && (
        <section aria-label="OMRChecker detected responses">
          <div className="section-heading">
            <h3>
              {reading.fields.length} fields read ·{" "}
              {reading.fields.filter((f) => f.uncertain).length} need review
            </h3>
            <span className="muted">
              {reading.engine} · {reading.revision.slice(0, 7)}
            </span>
          </div>
          <p>{reading.notice}</p>
          {!!reading.identity?.length && (
            <div className="omr-identity-results">
              {reading.identity.map((field, i) => (
                <div key={field.kind}>
                  <Field
                    name={
                      field.kind === "student_name"
                        ? "Detected student name"
                        : "Detected roll number"
                    }
                  >
                    <input
                      value={field.text}
                      placeholder={field.reason || "Review this field"}
                      onChange={(e) =>
                        setReading({
                          ...reading,
                          identity: reading.identity!.map((old, j) =>
                            j === i
                              ? {
                                  ...old,
                                  text: e.target.value,
                                  uncertain: false,
                                  reason: "Manually reviewed",
                                }
                              : old,
                          ),
                        })
                      }
                    />
                  </Field>
                  <small
                    role="status"
                    className={field.uncertain ? "error" : "muted"}
                  >
                    {field.reason === "Manually reviewed"
                      ? field.reason
                      : field.uncertain
                        ? `Needs review · ${field.reason || "Check the detected text."}`
                        : `Recognized · ${Math.round(field.confidence)}% OCR confidence`}
                  </small>
                </div>
              ))}
              <p className="muted">
                Check these values against the student record. Text recognition
                can need correction, especially for handwriting.
              </p>
            </div>
          )}
          <div className="omrchecker-results">
            <div className="omrchecker-table">
              <table>
                <caption>Detected template fields</caption>
                <thead>
                  <tr>
                    <th>Field</th>
                    <th>Filled response</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {reading.fields.map((f) => (
                    <tr key={f.label}>
                      <th scope="row">{f.label}</th>
                      <td>
                        {f.selection.length
                          ? f.selection.join(" + ")
                          : f.value || "Blank"}
                      </td>
                      <td>
                        {f.uncertain ? (
                          <span title={f.reason}>Review · {f.reason}</span>
                        ) : (
                          "Detected"
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              <details>
                <summary>Combined template fields</summary>
                <dl>
                  {Object.entries(reading.responses).map(([name, value]) => (
                    <div key={name}>
                      <dt>{name}</dt>
                      <dd>{value || "Blank"}</dd>
                    </div>
                  ))}
                </dl>
              </details>
            </div>
            {overlay && (
              <div>
                {/* Private local blob; no public asset URL. */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={overlay}
                  alt="OMRChecker aligned sheet with detected bubbles marked"
                />
              </div>
            )}
          </div>
          <button
            type="button"
            onClick={() => {
              const { overlay: _, ...report } = reading;
              const url = URL.createObjectURL(
                new Blob([JSON.stringify(report, null, 2)], {
                  type: "application/json",
                }),
              );
              const link = document.createElement("a");
              link.href = url;
              link.download = "omrchecker-responses.json";
              link.click();
              setTimeout(() => URL.revokeObjectURL(url), 1000);
            }}
          >
            Download detected answers
          </button>
        </section>
      )}
    </section>
  );
}
