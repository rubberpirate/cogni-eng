"use client";
import { useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { FormField as Field } from "./form-field";
type Q = A<"AcademicQuestion">;
export function QuestionEditor({
  initial,
  api,
  campus,
  course,
  busy,
  onSave,
  onCancel,
  requireSolution = true,
  simple = false,
}: {
  requireSolution?: boolean;
  simple?: boolean;
  initial: Q;
  api: EducationClient;
  campus: string;
  course: string;
  busy: boolean;
  onSave: (q: Q) => void;
  onCancel: () => void;
}) {
  const [q, setQ] = useState(initial);
  const [error, setError] = useState("");
  const [uploading, setUploading] = useState(false);
  const [rubric, setRubric] = useState<A<"AcademicCriterion">[]>(
    q.rubric?.length
      ? q.rubric
      : [{ id: "reason", description: "Correct reasoning", marks: q.marks }],
  );
  const [numeric, setNumeric] = useState<A<"AcademicNumericRule">>(
    q.numeric ?? { values: [], integer: false },
  );
  const [accepted, setAccepted] = useState(
    q.numeric?.values?.join(" | ") ?? "",
  );
  const [useRange, setUseRange] = useState(!!q.numeric?.min);
  const metadataFields = (
    <>
      <Field name="Difficulty">
        <select
          value={q.difficulty}
          onChange={(e) => setQ({ ...q, difficulty: e.target.value })}
        >
          {["introductory", "standard", "challenging"].map((d) => (
            <option key={d}>{d}</option>
          ))}
        </select>
      </Field>
      <Field name="Question language">
        <select
          value={q.language}
          onChange={(e) => setQ({ ...q, language: e.target.value })}
        >
          <option value="en">English</option>
          <option value="hi">हिन्दी</option>
        </select>
      </Field>
      <Field name="Source / provenance">
        <input
          value={q.provenance}
          required
          onChange={(e) => setQ({ ...q, provenance: e.target.value })}
        />
      </Field>
      <Field name="Usage rights">
        <input
          value={q.license}
          required
          onChange={(e) => setQ({ ...q, license: e.target.value })}
        />
      </Field>
    </>
  );
  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        const written = ["short", "extended"].includes(q.type);
        onSave({
          ...q,
          options: ["single", "multiple"].includes(q.type) ? q.options : [],
          key: ["single", "multiple"].includes(q.type) ? q.key : [],
          numeric:
            q.type === "numerical"
              ? {
                  ...numeric,
                  values: accepted
                    .split("|")
                    .map((v) => v.trim())
                    .filter(Boolean),
                  min: useRange ? numeric.min : undefined,
                  max: useRange ? numeric.max : undefined,
                }
              : undefined,
          rubric: written ? rubric : [],
        });
      }}
    >
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <Field name="Question type">
        <select
          value={q.type}
          onChange={(e) => setQ({ ...q, type: e.target.value })}
        >
          {[
            ["single", "Single choice"],
            ["multiple", "Multiple choices"],
            ["numerical", "Number answer"],
            ["short", "Short written answer"],
            ["extended", "Extended written answer"],
          ].map(([value, label]) => (
            <option key={value} value={value}>
              {label}
            </option>
          ))}
        </select>
      </Field>
      <Field name="Question wording">
        <textarea
          required
          minLength={3}
          value={q.stem}
          onChange={(e) => setQ({ ...q, stem: e.target.value })}
        />
      </Field>
      {["single", "multiple"].includes(q.type) && (
        <>
          <Field name="Number of answer options">
            <select
              value={q.options?.length ?? 4}
              onChange={(e) => {
                const n = Number(e.target.value);
                setQ({
                  ...q,
                  options: Array.from(
                    { length: n },
                    (_, i) =>
                      q.options?.[i] ?? {
                        id: String.fromCharCode(65 + i),
                        text: "",
                      },
                  ),
                  key: q.key?.filter((k) => k.charCodeAt(0) < 65 + n) ?? [],
                });
              }}
            >
              {[2, 3, 4, 5, 6, 7, 8].map((n) => (
                <option key={n}>{n}</option>
              ))}
            </select>
          </Field>
          {q.options?.map((o, i) => (
            <div className="option-editor" key={o.id}>
              <Field name={`Option ${o.id}`}>
                <input
                  required
                  value={o.text}
                  onChange={(e) =>
                    setQ({
                      ...q,
                      options: q.options!.map((v, j) =>
                        j === i ? { ...v, text: e.target.value } : v,
                      ),
                    })
                  }
                />
              </Field>
              <label className="check">
                <input
                  type={q.type === "single" ? "radio" : "checkbox"}
                  name="question-key"
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
                            : (q.key?.filter((k) => k !== o.id) ?? []),
                    })
                  }
                />
                Correct answer {o.id}
              </label>
            </div>
          ))}
        </>
      )}
      {q.type === "numerical" && (
        <section className="answer-rule-editor">
          <h3>Accepted numerical answer</h3>
          <Field name="Accepted answer(s)">
            <input
              required={!useRange}
              inputMode="decimal"
              value={accepted}
              onChange={(e) => setAccepted(e.target.value)}
              placeholder="For example, 0.3"
            />
          </Field>
          <p className="muted">
            Use a dot for decimals. Separate alternative accepted values with |.
          </p>
          <label className="check">
            <input
              type="checkbox"
              checked={useRange}
              onChange={(e) => setUseRange(e.target.checked)}
            />
            Accept any number in a range
          </label>
          {useRange && (
            <div className="academic-columns">
              <Field name="Lowest accepted value">
                <input
                  required
                  inputMode="decimal"
                  value={numeric.min ?? ""}
                  onChange={(e) =>
                    setNumeric({ ...numeric, min: e.target.value })
                  }
                />
              </Field>
              <Field name="Highest accepted value">
                <input
                  required
                  inputMode="decimal"
                  value={numeric.max ?? ""}
                  onChange={(e) =>
                    setNumeric({ ...numeric, max: e.target.value })
                  }
                />
              </Field>
            </div>
          )}
          <label className="check">
            <input
              type="checkbox"
              checked={numeric.integer}
              onChange={(e) =>
                setNumeric({ ...numeric, integer: e.target.checked })
              }
            />
            Accept whole numbers only
          </label>
          <details>
            <summary>Allow a tolerance around accepted values</summary>
            <Field name="Allowed absolute difference">
              <input
                inputMode="decimal"
                value={numeric.absolute ?? ""}
                onChange={(e) =>
                  setNumeric({ ...numeric, absolute: e.target.value })
                }
                placeholder="0 means exact"
              />
            </Field>
            <Field name="Allowed relative difference (0.01 = 1%)">
              <input
                inputMode="decimal"
                value={numeric.relative ?? ""}
                onChange={(e) =>
                  setNumeric({ ...numeric, relative: e.target.value })
                }
              />
            </Field>
          </details>
        </section>
      )}
      {["short", "extended"].includes(q.type) && (
        <section className="answer-rule-editor">
          <h3>How should this written answer be marked?</h3>
          <p>
            Describe each criterion and its available marks. Criterion marks
            must add up to the question's total marks.
          </p>
          {rubric.map((c, i) => (
            <div className="rubric-row" key={c.id}>
              <Field name={`Criterion ${i + 1}: what earns marks?`}>
                <textarea
                  required
                  value={c.description}
                  onChange={(e) =>
                    setRubric(
                      rubric.map((r, j) =>
                        j === i ? { ...r, description: e.target.value } : r,
                      ),
                    )
                  }
                />
              </Field>
              <Field name={`Marks for criterion ${i + 1}`}>
                <input
                  required
                  inputMode="decimal"
                  value={c.marks}
                  onChange={(e) =>
                    setRubric(
                      rubric.map((r, j) =>
                        j === i ? { ...r, marks: e.target.value } : r,
                      ),
                    )
                  }
                />
              </Field>
              <button
                type="button"
                disabled={rubric.length === 1}
                onClick={() => setRubric(rubric.filter((_, j) => j !== i))}
              >
                Remove criterion {i + 1}
              </button>
            </div>
          ))}
          <button
            type="button"
            disabled={rubric.length >= 20}
            onClick={() =>
              setRubric([
                ...rubric,
                { id: crypto.randomUUID(), description: "", marks: "1" },
              ])
            }
          >
            Add marking criterion
          </button>
        </section>
      )}
      <div className="academic-columns">
        <Field name="Marks">
          <input
            value={q.marks}
            inputMode="decimal"
            required
            onChange={(e) => setQ({ ...q, marks: e.target.value })}
          />
        </Field>
        <Field name="Wrong response penalty">
          <input
            value={q.penalty}
            inputMode="decimal"
            required
            onChange={(e) => setQ({ ...q, penalty: e.target.value })}
          />
        </Field>
      </div>
      <Field name="Worked solution / marking guidance">
        <textarea
          value={q.solution}
          required={requireSolution}
          onChange={(e) => setQ({ ...q, solution: e.target.value })}
        />
      </Field>
      {simple ? (
        <details>
          <summary>Question details & source</summary>
          {metadataFields}
        </details>
      ) : (
        metadataFields
      )}
      <details>
        <summary>Attach a question diagram</summary>
        <Field name="Diagram file">
          <input
            type="file"
            accept="image/png,image/jpeg"
            disabled={uploading}
            onChange={async (e) => {
              const file = e.target.files?.[0];
              if (!file) return;
              setUploading(true);
              try {
                const f = new FormData();
                f.set("file", file);
                f.set("campus_id", campus);
                f.set("course_id", course);
                f.set("purpose", "question_image");
                const asset = await api.asset(f);
                setQ((q) => ({ ...q, asset_id: asset.id }));
                setError("");
              } catch (e) {
                setError(
                  e instanceof Error ? e.message : "Image upload interrupted",
                );
              } finally {
                setUploading(false);
              }
            }}
          />
        </Field>
        {q.asset_id && <p>Private diagram attached.</p>}
        <Field name="Diagram alternative text">
          <textarea
            value={q.alt ?? ""}
            onChange={(e) => setQ({ ...q, alt: e.target.value })}
          />
        </Field>
      </details>
      <div className="actions">
        <button disabled={busy || uploading}>Save question draft</button>
        <button type="button" onClick={onCancel}>
          Cancel editor
        </button>
      </div>
    </form>
  );
}
