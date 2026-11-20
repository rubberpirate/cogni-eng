"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import type { A, EducationClient } from "../../../clients/academic";
import { FormField as Field } from "./form-field";
import { QuestionEditor } from "./question-editor";
import { KeyImportPanel, MarkingRules } from "./import-review";
import { TaskProgress } from "./task-progress";
import { WorkIcon } from "./workspace-art";
type Draft = A<"AcademicAssessmentDraft">;
type Spec = A<"AcademicDraftSpec">;
type Q = A<"AcademicQuestion">;
type Topic = A<"AcademicTopic">;
type Job = A<"AcademicJob">;
type Props = {
  api: EducationClient;
  courses: A<"AcademicCourse">[];
  topics: Topic[];
  papers: A<"AcademicPaperRecord">[];
  campus: string;
  active: boolean;
  onRefresh: () => Promise<void>;
  onResults: (p: A<"AcademicPaperRecord">) => void;
};
const localDate = (d: string) => {
  const t = new Date(d);
  return new Date(t.getTime() - t.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16);
};
const freshQuestion = (topic: string): Q => ({
  topic_id: topic,
  type: "single",
  stem: "",
  options: ["A", "B", "C", "D"].map((id) => ({ id, text: "" })),
  key: [],
  rubric: [],
  solution: "",
  marks: "4",
  penalty: "1",
  difficulty: "standard",
  language: "en",
  provenance: "Faculty-authored question",
  license: "Institution-owned",
});
const freshSpec = (method: string): Spec => ({
  title: "",
  method,
  topic_id: "",
  exam: "IIT-JEE",
  starts_at: new Date(Date.now() + 86400000).toISOString(),
  duration_minutes: 180,
  instructions:
    "Read every question carefully. Blank answers receive zero marks.",
  learners: [],
  questions: [],
  import_id: "",
  job_ids: [],
  applied_jobs: [],
  source_applied: false,
});
export function AssessmentWorkspace({
  api,
  courses,
  topics,
  papers,
  campus,
  active,
  onRefresh,
  onResults,
}: Props) {
  const [drafts, setDrafts] = useState<Draft[]>([]),
    [draft, setDraft] = useState<Draft | null>(null),
    [method, setMethod] = useState("");
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [course, setCourse] = useState(""),
    [spec, setSpec] = useState<Spec>(freshSpec("manual"));
  const [roster, setRoster] = useState<A<"AcademicRecipient">[]>([]);
  const [subject, setSubject] = useState("Physics"),
    [topicName, setTopicName] = useState("");
  const [query, setQuery] = useState("");
  const loaded = useRef("");
  const loadDrafts = useCallback(
    async () => setDrafts(await api.list("drafts", { campus })),
    [api, campus],
  );
  useEffect(() => {
    void loadDrafts().catch((e) => setError(e.message));
  }, [loadDrafts]);
  useEffect(() => {
    let valid = true;
    setRoster([]);
    if (course)
      void api
        .courseLearners(course)
        .then((rows) => {
          if (valid) {
            setRoster(rows);
            if (!draft)
              setSpec((s) => ({ ...s, learners: rows.map((r) => r.id) }));
          }
        })
        .catch((e) => {
          if (valid) setError(e.message);
        });
    return () => {
      valid = false;
    };
  }, [api, course, draft?.id]);
  const open = useCallback((d: Draft) => {
    loaded.current = d.id;
    setDraft(d);
    setCourse(d.course_id);
    setSpec(d.spec);
    setMethod(d.spec.method);
    setError("");
    window.location.hash = `education/draft?draft=${d.id}`;
  }, []);
  useEffect(() => {
    let alive = true;
    const restore = (event?: Event) => {
      const id = new URLSearchParams(window.location.hash.split("?")[1]).get(
        "draft",
      );
      if (!id) {
        loaded.current = "";
        setDraft(null);
        setMethod("");
        setCourse("");
        setSpec(freshSpec("manual"));
      } else if (!event || id !== loaded.current) {
        loaded.current = id;
        void api
          .detail("drafts", id)
          .then((d) => {
            if (alive && loaded.current === id) open(d);
          })
          .catch((e) => {
            if (alive) setError(e.message);
          });
      }
    };
    restore();
    window.addEventListener("hashchange", restore);
    return () => {
      alive = false;
      window.removeEventListener("hashchange", restore);
    };
  }, [api, open]);
  const createTopic = async (
    c: A<"AcademicCourse">,
    sub: string,
    name: string,
  ) => {
    const available = await api.list("topics", { campus: c.campus_id });
    const existing = available.find(
      (t) =>
        t.course_id === c.id &&
        t.subject.toLowerCase() === sub.toLowerCase() &&
        t.name.toLowerCase() === name.toLowerCase(),
    );
    return (
      existing ??
      (await api.topics({
        id: "",
        campus_id: c.campus_id,
        course_id: c.id,
        subject: sub,
        name,
        outcome: `Assess understanding and application of ${name}.`,
        language: "en",
      }))
    );
  };
  const start = async () => {
    setBusy(true);
    setError("");
    try {
      const c = courses.find((c) => c.id === course);
      if (!c) throw Error("Choose a student batch.");
      const t = await createTopic(c, subject, topicName || spec.title);
      const d = await api.saveDraft({
        campus_id: c.campus_id,
        course_id: c.id,
        version: 0,
        spec: { ...spec, topic_id: t.id, method },
      });
      open(d);
      await loadDrafts();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not save draft.");
    } finally {
      setBusy(false);
    }
  };
  const clone = async (p: A<"AcademicPaperRecord">) => {
    setBusy(true);
    setError("");
    try {
      const [students, assigned] = await Promise.all([
        api.courseLearners(p.course_id),
        api.list("assignments", { paper: p.id }),
      ]);
      const questions = p.spec.sections.flatMap((s) => s.questions);
      const next = await api.saveDraft({
        campus_id: p.campus_id,
        course_id: p.course_id,
        version: 0,
        spec: {
          ...freshSpec("manual"),
          title: p.spec.title + " — revision",
          exam: ["IIT-JEE", "NEET", "Other assessment"].includes(
            p.spec.sections[0]?.title,
          )
            ? p.spec.sections[0].title
            : "IIT-JEE",
          topic_id: questions[0]?.topic_id ?? "",
          duration_minutes: p.spec.duration_minutes,
          instructions: p.spec.instructions,
          learners: students
            .filter((s) => assigned.some((a) => a.learner_id === s.id))
            .map((s) => s.id),
          questions,
        },
      });
      open(next);
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Could not create an editable revision.",
      );
    } finally {
      setBusy(false);
    }
  };
  if (draft)
    return (
      <DraftEditor
        key={draft.id}
        api={api}
        initial={draft}
        courses={courses}
        topics={topics}
        createTopic={createTopic}
        onExit={() => {
          setDraft(null);
          setMethod("");
          loaded.current = "";
          window.location.hash = active
            ? "education/active"
            : "education/draft";
          void loadDrafts();
          void onRefresh();
        }}
      />
    );
  return (
    <div className="assessment-workspace">
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {active ? (
        <section className="card">
          <div className="section-heading">
            <h2>Active assessments</h2>
            <span className="muted">
              Published papers and assigned students
            </span>
          </div>
          <Field name="Search assessments">
            <input
              type="search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Assessment name or reference"
            />
          </Field>
          {papers
            .filter(
              (p) =>
                p.state === "published" &&
                `${p.spec.title} ${p.reference}`
                  .toLowerCase()
                  .includes(query.toLowerCase()),
            )
            .map((p) => (
              <article className="academic-row" key={p.id}>
                <div>
                  <strong>{p.spec.title}</strong>
                  <p className="muted">
                    {p.reference} ·{" "}
                    {p.spec.sections.reduce(
                      (n, s) => n + s.questions.length,
                      0,
                    )}{" "}
                    questions · {p.maximum} marks
                  </p>
                </div>
                <div className="actions">
                  <span className="muted">{p.spec.duration_minutes} min</span>
                  <button disabled={busy} onClick={() => void clone(p)}>
                    Edit as new revision
                  </button>
                  <button onClick={() => onResults(p)}>View results</button>
                </div>
              </article>
            ))}
          {!papers.some((p) => p.state === "published") && (
            <p className="empty-message">
              Publish a reviewed draft to see it here.
            </p>
          )}
          <p className="muted">
            Editing creates a new draft. Students' existing papers and results
            retain their original version.
          </p>
        </section>
      ) : (
        <>
          {!method ? (
            <>
              <section
                className="draft-methods"
                aria-label="Choose how to draft your assessment"
              >
                {(
                  [
                    [
                      "manual",
                      "curriculum",
                      "Enter manually",
                      "Write each question, its options and answer.",
                    ],
                    [
                      "upload",
                      "import",
                      "Upload question paper",
                      "Read a PDF or photo, then review the extracted paper.",
                    ],
                    [
                      "ai",
                      "generate",
                      "Generate with AI",
                      "Build question groups by subject, topic and difficulty.",
                    ],
                  ] as const
                ).map(([id, icon, title, detail]) => (
                  <button
                    key={id}
                    className={`draft-method-choice choice-${id}`}
                    onClick={() => {
                      setMethod(id);
                      setSpec(freshSpec(id));
                    }}
                  >
                    <WorkIcon name={icon} />
                    <strong>{title}</strong>
                    <span>{detail}</span>
                    <span aria-hidden="true">↗</span>
                  </button>
                ))}
              </section>
              <details className="saved-drafts">
                <summary>
                  Resume a saved draft (
                  {drafts.filter((d) => d.state === "draft").length})
                </summary>
                {drafts
                  .filter((d) => d.state === "draft")
                  .map((d) => (
                    <button
                      className="paper-choice"
                      key={d.id}
                      onClick={() => open(d)}
                    >
                      {d.spec.title}
                      <small>
                        {d.spec.questions.length} questions ·{" "}
                        {new Date(d.spec.starts_at).toLocaleString()}
                      </small>
                    </button>
                  ))}
                {!drafts.some((d) => d.state === "draft") && (
                  <p>No saved drafts yet.</p>
                )}
              </details>
            </>
          ) : (
            <section className="card draft-setup">
              <button
                className="dashboard-text-link"
                onClick={() => setMethod("")}
              >
                ← Change creation method
              </button>
              <h2>Assessment details</h2>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  void start();
                }}
              >
                <Field name="Assessment name">
                  <input
                    value={spec.title}
                    onChange={(e) =>
                      setSpec({ ...spec, title: e.target.value })
                    }
                    required
                    minLength={3}
                    maxLength={200}
                  />
                </Field>
                <div className="academic-columns">
                  <Field name="Student batch">
                    <select
                      value={course}
                      required
                      onChange={(e) => setCourse(e.target.value)}
                    >
                      <option value="">Choose your batch / course</option>
                      {courses.map((c) => (
                        <option key={c.id} value={c.id}>
                          {c.name} · {c.academic_year}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field name="Assessment date and time">
                    <input
                      type="datetime-local"
                      required
                      value={localDate(spec.starts_at)}
                      onChange={(e) =>
                        e.target.value &&
                        setSpec({
                          ...spec,
                          starts_at: new Date(e.target.value).toISOString(),
                        })
                      }
                    />
                  </Field>
                  <Field name="Duration (minutes)">
                    <input
                      type="number"
                      min={1}
                      max={480}
                      required
                      value={spec.duration_minutes}
                      onChange={(e) =>
                        setSpec({
                          ...spec,
                          duration_minutes: Number(e.target.value),
                        })
                      }
                    />
                  </Field>
                  <Field name="Paper type">
                    <select
                      value={spec.exam}
                      onChange={(e) =>
                        setSpec({ ...spec, exam: e.target.value })
                      }
                    >
                      <option>IIT-JEE</option>
                      <option>NEET</option>
                      <option>Other assessment</option>
                    </select>
                  </Field>
                  <Field name="Main subject">
                    <input
                      required
                      minLength={2}
                      value={subject}
                      onChange={(e) => setSubject(e.target.value)}
                      list="draft-subjects"
                    />
                  </Field>
                  <datalist id="draft-subjects">
                    <option>Physics</option>
                    <option>Chemistry</option>
                    <option>Mathematics</option>
                    <option>Biology</option>
                  </datalist>
                  <Field name="Main topic">
                    <input
                      value={topicName}
                      onChange={(e) => setTopicName(e.target.value)}
                      placeholder="Use assessment name if mixed topics"
                    />
                  </Field>
                </div>
                <Field name="Paper instructions">
                  <textarea
                    value={spec.instructions}
                    onChange={(e) =>
                      setSpec({ ...spec, instructions: e.target.value })
                    }
                  />
                </Field>
                <details>
                  <summary>
                    Students in this batch ({spec.learners.length} selected)
                  </summary>
                  <p className="muted">
                    Optional for online assignments. Paper answer sheets can be
                    checked by registration number or name without student
                    accounts.
                  </p>
                  {roster.map((r) => (
                    <label className="check" key={r.id}>
                      <input
                        type="checkbox"
                        checked={spec.learners.includes(r.id)}
                        onChange={(e) =>
                          setSpec({
                            ...spec,
                            learners: e.target.checked
                              ? [...spec.learners, r.id]
                              : spec.learners.filter((id) => id !== r.id),
                          })
                        }
                      />
                      {r.name} · {r.reference}
                    </label>
                  ))}
                </details>
                <button className="primary" disabled={busy || !course}>
                  {busy ? "Saving setup…" : "Continue to questions"}
                </button>
              </form>
            </section>
          )}
        </>
      )}
    </div>
  );
}

function DraftEditor({
  api,
  initial,
  courses,
  topics,
  createTopic,
  onExit,
}: {
  api: EducationClient;
  initial: Draft;
  courses: A<"AcademicCourse">[];
  topics: Topic[];
  createTopic: (
    c: A<"AcademicCourse">,
    subject: string,
    name: string,
  ) => Promise<Topic>;
  onExit: () => void;
}) {
  const [draft, setDraft] = useState(initial),
    [spec, setSpec] = useState(initial.spec),
    [dirty, setDirty] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState("");
  const current = useRef(draft);
  current.current = draft;
  const [source, setSource] = useState(
    initial.spec.questions.length ||
      initial.spec.source_applied ||
      initial.spec.applied_jobs?.length
      ? ""
      : initial.spec.method,
  );
  const [edit, setEdit] = useState<{ index: number; question: Q } | null>(null),
    [reviewed, setReviewed] = useState(false);
  const [preview, setPreview] = useState<A<"AcademicImport"> | null>(null),
    [jobs, setJobs] = useState<Job[]>([]);
  const [extractionJob, setExtractionJob] = useState<Job | null>(null);
  const [groups, setGroups] = useState([
    { subject: "Physics", topic: "", count: 3, difficulty: "standard" },
  ]);
  const [newTopic, setNewTopic] = useState(initial.spec.topic_id),
    [extraSubject, setExtraSubject] = useState(""),
    [extraTopic, setExtraTopic] = useState("");
  const [roster, setRoster] = useState<A<"AcademicRecipient">[]>([]);
  const consumed = useRef(new Set<string>());
  const c = courses.find((c) => c.id === draft.course_id);
  const change = (next: Spec) => {
    setSpec(next);
    setDirty(true);
    setReviewed(false);
    setNotice("");
  };
  const save = async (next = spec) => {
    const saved = await api.saveDraft(
      {
        campus_id: draft.campus_id,
        course_id: draft.course_id,
        version: current.current.version,
        spec: next,
      },
      draft.id,
    );
    current.current = saved;
    setDraft(saved);
    setSpec(saved.spec);
    setDirty(false);
    return saved;
  };
  const run = async (fn: () => Promise<void>) => {
    setBusy(true);
    setError("");
    try {
      await fn();
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Could not complete this step. Your saved draft is retained.",
      );
    } finally {
      setBusy(false);
    }
  };
  useEffect(() => {
    const warn = (e: BeforeUnloadEvent) => {
      if (dirty || edit) {
        e.preventDefault();
        e.returnValue = "";
      }
    };
    const guard = (e: MouseEvent) => {
      if (
        (dirty || edit) &&
        (e.target as Element)?.closest?.(
          ".rail button,.context-bar button,.wordmark",
        )
      ) {
        e.preventDefault();
        e.stopImmediatePropagation();
        setError("Save your paper changes before navigating away.");
      }
    };
    document.addEventListener("click", guard, true);
    window.addEventListener("beforeunload", warn);
    return () => {
      document.removeEventListener("click", guard, true);
      window.removeEventListener("beforeunload", warn);
    };
  }, [dirty, edit]);
  useEffect(() => {
    let alive = true;
    void api
      .courseLearners(draft.course_id)
      .then((r) => {
        if (alive) setRoster(r);
      })
      .catch((e) => {
        if (alive) setError(e.message);
      });
    if (initial.spec.import_id)
      void api
        .detail("imports", initial.spec.import_id)
        .then((p) => {
          if (alive) setPreview(p);
        })
        .catch((e) => {
          if (alive) setError(e.message);
        });
    return () => {
      alive = false;
    };
  }, [api, draft.course_id, initial.spec.import_id]);
  const syncJobs = useCallback(async () => {
    const spec = current.current.spec;
    const ids = spec.job_ids.filter(
      (id) => !(spec.applied_jobs ?? []).includes(id),
    );
    if (!ids.length) {
      setJobs([]);
      return;
    }
    setJobs(await Promise.all(ids.map((id) => api.detail("jobs", id))));
  }, [api]);
  useEffect(() => {
    void syncJobs().catch((e) => setError(e.message));
    const timer = setInterval(
      () => void syncJobs().catch((e) => setError(e.message)),
      2500,
    );
    return () => clearInterval(timer);
  }, [syncJobs]);
  useEffect(() => {
    if (!preview || preview.state !== "extracting") return;
    let alive = true;
    const timer = setInterval(
      () =>
        void api
          .detail("imports", preview.id)
          .then((p) => {
            if (alive) setPreview(p);
            return api.list("jobs").then((list) => {
              if (alive)
                setExtractionJob(
                  list.find(
                    (j) =>
                      j.kind === "extract_document" &&
                      j.resource_id === preview.id,
                  ) ?? null,
                );
            });
          })
          .catch((e) => {
            if (alive) setError(e.message);
          }),
      2500,
    );
    return () => {
      alive = false;
      clearInterval(timer);
    };
  }, [api, preview?.id, preview?.state]);
  // Persist source adoption separately so deleting all questions never reimports them.
  useEffect(() => {
    const p = preview;
    if (
      !p ||
      p.state !== "preview" ||
      !p.rows.length ||
      spec.questions.length ||
      spec.source_applied ||
      consumed.current.has(p.id)
    )
      return;
    consumed.current.add(p.id);
    void run(async () => {
      await save({
        ...spec,
        source_applied: true,
        questions: p.rows.map((r) => ({
          ...r.question,
          id: `source:${p.id}:${r.index}`,
          origin: "imported",
        })),
      });
      setSource("");
      setNotice(
        "Questions extracted. Check the wording, answers and marking scheme below.",
      );
    });
  }, [preview?.id, preview?.state]);
  useEffect(() => {
    if (
      !jobs.length ||
      jobs.some((j) => ["queued", "running"].includes(j.state)) ||
      consumed.current.has(jobs.map((j) => j.id).join("|"))
    )
      return;
    const completed = jobs.filter((j) => j.state === "completed");
    if (!completed.length) return;
    const token = jobs.map((j) => j.id).join("|");
    consumed.current.add(token);
    void run(async () => {
      const added = completed.flatMap((j) =>
        ((j.output as { questions?: Q[] }).questions ?? []).map((q) => ({
          ...q,
          origin: "generated",
        })),
      );
      if (added.length) {
        setReviewed(false);
        await save({
          ...spec,
          applied_jobs: [
            ...new Set([
              ...(spec.applied_jobs ?? []),
              ...completed.map((j) => j.id),
            ]),
          ],
          questions: [...spec.questions, ...added],
        });
        setSource("");
        setJobs(
          jobs.filter((j) => !completed.some((done) => done.id === j.id)),
        );
        setNotice(
          "Generated drafts added. Verify answers and solutions before publishing.",
        );
      }
    });
  }, [jobs]);
  const processing =
    busy ||
    preview?.state === "extracting" ||
    jobs.some((j) => ["queued", "running"].includes(j.state));
  const upload = async (file: File) => {
    await run(async () => {
      await save();
      const f = new FormData();
      f.set("file", file);
      f.set("engine", "tesseract");
      f.set("topic_id", spec.topic_id);
      f.set("provenance", "Faculty uploaded assessment paper");
      f.set("license", "Institution-owned");
      const p = await api.uploadImport(f);
      await save({ ...current.current.spec, import_id: p.id });
      setPreview(p);
      if (p.state === "preview" && p.rows.length) {
        await save({
          ...current.current.spec,
          source_applied: true,
          questions: p.rows.map((r) => ({
            ...r.question,
            id: `source:${p.id}:${r.index}`,
            origin: "imported",
          })),
        });
        consumed.current.add(p.id);
        setSource("");
      }
    });
  };
  const generate = async () => {
    await run(async () => {
      if (!c) throw Error("Batch unavailable.");
      if (spec.questions.length + groups.reduce((n, g) => n + g.count, 0) > 200)
        throw Error("Use at most 200 questions in one assessment.");
      await save();
      let next = current.current.spec;
      const ids = [...next.job_ids];
      for (const g of groups) {
        const t = await createTopic(c, g.subject, g.topic);
        const j = await api.generate({
          topic_id: t.id,
          count: g.count,
          difficulty: g.difficulty,
          exam: spec.exam,
        });
        ids.push(j.id);
        next = { ...next, job_ids: [...ids] };
        await save(next);
      }
      await syncJobs();
    });
  };
  const manual = async () => {
    await run(async () => {
      let t = newTopic;
      if (extraSubject || extraTopic) {
        if (!c || extraSubject.length < 2 || extraTopic.length < 2)
          throw Error("Enter both subject and topic.");
        t = (await createTopic(c, extraSubject, extraTopic)).id;
      }
      if (!t) throw Error("Choose a topic.");
      setEdit({ index: -1, question: freshQuestion(t) });
    });
  };
  const exit = () => {
    if (dirty || edit) {
      setError(
        "Save or cancel the current question and save paper changes before leaving.",
      );
      return;
    }
    onExit();
  };
  return (
    <div className="assessment-editor">
      <div className="editor-toolbar">
        <button onClick={exit}>← Assessments</button>
        <div>
          <strong>{spec.title}</strong>
          <span className="muted">
            {draft.state === "published"
              ? "Published"
              : dirty
                ? "Unsaved changes"
                : "Saved draft"}{" "}
            · {spec.questions.length} questions
          </span>
        </div>
        {draft.state === "draft" && (
          <button
            className="primary"
            disabled={busy}
            onClick={() =>
              void run(async () => {
                await save();
                setNotice("Paper changes saved.");
              })
            }
          >
            Save changes
          </button>
        )}
      </div>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {notice && (
        <p role="status" className="notice">
          {notice}
        </p>
      )}
      {draft.state === "published" ? (
        <section className="card">
          <h2>Assessment published and assigned</h2>
          <p>{spec.title} is ready for the selected students.</p>
          <button onClick={onExit}>Done</button>
        </section>
      ) : (
        <>
          <details className="card paper-details">
            <summary>Assessment details & students</summary>
            <div className="academic-columns">
              <Field name="Assessment name">
                <input
                  value={spec.title}
                  onChange={(e) => change({ ...spec, title: e.target.value })}
                />
              </Field>
              <Field name="Assessment date and time">
                <input
                  type="datetime-local"
                  value={localDate(spec.starts_at)}
                  onChange={(e) =>
                    e.target.value &&
                    change({
                      ...spec,
                      starts_at: new Date(e.target.value).toISOString(),
                    })
                  }
                />
              </Field>
              <Field name="Duration (minutes)">
                <input
                  type="number"
                  min={1}
                  max={480}
                  value={spec.duration_minutes}
                  onChange={(e) =>
                    change({
                      ...spec,
                      duration_minutes: Number(e.target.value),
                    })
                  }
                />
              </Field>
              <Field name="Paper instructions">
                <textarea
                  value={spec.instructions}
                  onChange={(e) =>
                    change({ ...spec, instructions: e.target.value })
                  }
                />
              </Field>
            </div>
            <p>
              {c?.name} · {spec.learners.length} selected students
            </p>
            <div className="actions">
              <button
                onClick={() =>
                  change({ ...spec, learners: roster.map((r) => r.id) })
                }
              >
                Select all active students
              </button>
              <button onClick={() => change({ ...spec, learners: [] })}>
                Clear student selection
              </button>
            </div>
            {roster.map((r) => (
              <label key={r.id} className="check">
                <input
                  type="checkbox"
                  checked={spec.learners.includes(r.id)}
                  onChange={(e) =>
                    change({
                      ...spec,
                      learners: e.target.checked
                        ? [...spec.learners, r.id]
                        : spec.learners.filter((id) => id !== r.id),
                    })
                  }
                />
                {r.name} · {r.reference}
              </label>
            ))}
          </details>
          {source === "upload" && !preview && (
            <section className="card upload-only">
              <h2>Upload your question paper</h2>
              <label
                className="paper-drop-zone"
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => {
                  e.preventDefault();
                  const f = e.dataTransfer.files[0];
                  if (f) void upload(f);
                }}
              >
                <WorkIcon name="import" />
                <strong>Drop a question paper or choose a file</strong>
                <span>PDF, PNG or JPEG</span>
                <input
                  aria-label="Question paper file"
                  type="file"
                  accept="application/pdf,image/png,image/jpeg"
                  disabled={busy}
                  onChange={(e) => {
                    const f = e.target.files?.[0];
                    if (f) void upload(f);
                  }}
                />
              </label>
            </section>
          )}
          {preview?.state === "extracting" && (
            <TaskProgress
              state={extractionJob?.state ?? "running"}
              title="Reading your paper with OCR and local AI"
              detail={
                extractionJob?.error ||
                "The uploaded source is saved. Questions and options will open in the editor after processing."
              }
            />
          )}
          {preview?.state === "extracting" && (
            <button
              disabled={busy}
              onClick={() =>
                void run(async () => {
                  const all = await api.list("jobs");
                  const j = all.find(
                    (j) =>
                      j.resource_id === preview.id &&
                      j.kind === "extract_document",
                  );
                  if (
                    j &&
                    ["failed", "configuration_required", "cancelled"].includes(
                      j.state,
                    )
                  )
                    await api.jobAction(j.id, "retry");
                  setExtractionJob(j ?? null);
                })
              }
            >
              Check processing / retry
            </button>
          )}
          {extractionJob &&
            ["queued", "running"].includes(extractionJob.state) && (
              <button
                disabled={busy}
                onClick={() =>
                  void run(async () => {
                    await api.jobAction(extractionJob.id, "cancel");
                    setExtractionJob(
                      await api.detail("jobs", extractionJob.id),
                    );
                  })
                }
              >
                Cancel paper processing
              </button>
            )}
          {preview?.asset_id && (
            <a
              href={api.url("assets/" + preview.asset_id)}
              target="_blank"
              rel="noreferrer"
            >
              Open original question paper
            </a>
          )}
          {preview?.raw_text && (
            <details>
              <summary>Source transcription</summary>
              <pre className="source-transcription">{preview.raw_text}</pre>
            </details>
          )}
          {source === "ai" && (
            <section className="card">
              <h2>Build your question groups</h2>
              <p className="muted">
                Each row adds a subject, topic, count and difficulty to this
                paper.
              </p>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  void generate();
                }}
              >
                {groups.map((g, i) => (
                  <fieldset className="generation-group" key={i}>
                    <legend>Question group {i + 1}</legend>
                    <Field name={`Group ${i + 1} subject`}>
                      <input
                        required
                        minLength={2}
                        value={g.subject}
                        onChange={(e) =>
                          setGroups(
                            groups.map((x, j) =>
                              j === i ? { ...x, subject: e.target.value } : x,
                            ),
                          )
                        }
                      />
                    </Field>
                    <Field name={`Group ${i + 1} topic or topic range`}>
                      <input
                        required
                        minLength={2}
                        value={g.topic}
                        placeholder="e.g. Electrostatics to capacitance"
                        onChange={(e) =>
                          setGroups(
                            groups.map((x, j) =>
                              j === i ? { ...x, topic: e.target.value } : x,
                            ),
                          )
                        }
                      />
                    </Field>
                    <Field name={`Group ${i + 1} question count`}>
                      <input
                        type="number"
                        min={1}
                        max={20}
                        required
                        value={g.count}
                        onChange={(e) =>
                          setGroups(
                            groups.map((x, j) =>
                              j === i
                                ? { ...x, count: Number(e.target.value) }
                                : x,
                            ),
                          )
                        }
                      />
                    </Field>
                    <Field name={`Group ${i + 1} difficulty`}>
                      <select
                        value={g.difficulty}
                        onChange={(e) =>
                          setGroups(
                            groups.map((x, j) =>
                              j === i
                                ? { ...x, difficulty: e.target.value }
                                : x,
                            ),
                          )
                        }
                      >
                        <option value="introductory">Easy</option>
                        <option value="standard">Medium</option>
                        <option value="challenging">Hard</option>
                      </select>
                    </Field>
                    <button
                      type="button"
                      disabled={groups.length === 1}
                      onClick={() =>
                        setGroups(groups.filter((_, j) => j !== i))
                      }
                    >
                      Remove group
                    </button>
                  </fieldset>
                ))}
                <div className="actions">
                  <button
                    type="button"
                    disabled={groups.length >= 10}
                    onClick={() =>
                      setGroups([
                        ...groups,
                        {
                          subject: "",
                          topic: "",
                          count: 5,
                          difficulty: "standard",
                        },
                      ])
                    }
                  >
                    Add another group
                  </button>
                  <button
                    className="primary"
                    disabled={
                      busy ||
                      jobs.some((j) => ["queued", "running"].includes(j.state))
                    }
                  >
                    Generate question groups
                  </button>
                </div>
              </form>
            </section>
          )}
          {jobs.map((j) => (
            <TaskProgress
              key={j.id}
              state={j.state}
              title={
                j.kind === "extract_document"
                  ? "Paper extraction"
                  : "Local question generation"
              }
              detail={
                j.error ||
                "Questions remain drafts until you review and publish the paper."
              }
              progress={j.progress}
            >
              {["failed", "configuration_required", "cancelled"].includes(
                j.state,
              ) && (
                <button
                  disabled={busy}
                  onClick={() =>
                    void run(async () => {
                      await api.jobAction(j.id, "retry");
                      await syncJobs();
                    })
                  }
                >
                  Retry this group
                </button>
              )}
              {["queued", "running"].includes(j.state) && (
                <button
                  onClick={() =>
                    void run(async () => {
                      await api.jobAction(j.id, "cancel");
                      await syncJobs();
                    })
                  }
                >
                  Cancel
                </button>
              )}
            </TaskProgress>
          ))}
          {source === "manual" && !edit && (
            <section className="card">
              <h2>Add a question manually</h2>
              <Field name="Question topic">
                <select
                  value={newTopic}
                  onChange={(e) => setNewTopic(e.target.value)}
                >
                  {[
                    {
                      id: spec.topic_id,
                      subject: "Assessment",
                      name: "Main topic",
                    },
                    ...topics.filter(
                      (t) =>
                        t.course_id === draft.course_id &&
                        t.id !== spec.topic_id,
                    ),
                  ].map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.subject} / {t.name}
                    </option>
                  ))}
                </select>
              </Field>
              <details>
                <summary>Use another subject or topic</summary>
                <Field name="New question subject">
                  <input
                    value={extraSubject}
                    onChange={(e) => setExtraSubject(e.target.value)}
                  />
                </Field>
                <Field name="New question topic">
                  <input
                    value={extraTopic}
                    onChange={(e) => setExtraTopic(e.target.value)}
                  />
                </Field>
              </details>
              <button
                disabled={busy}
                className="primary"
                onClick={() => void manual()}
              >
                Enter question
              </button>
            </section>
          )}
          {edit && (
            <section className="card">
              <h2>
                {edit.index < 0
                  ? "New question"
                  : `Edit question ${edit.index + 1}`}
              </h2>
              <QuestionEditor
                key={edit.index}
                requireSolution={false}
                simple
                initial={edit.question}
                api={api}
                campus={draft.campus_id}
                course={draft.course_id}
                busy={busy}
                onCancel={() => setEdit(null)}
                onSave={(q) =>
                  void run(async () => {
                    const questions =
                      edit.index < 0
                        ? [...spec.questions, q]
                        : spec.questions.map((old, i) =>
                            i === edit.index ? q : old,
                          );
                    await save({ ...spec, questions });
                    setReviewed(false);
                    setEdit(null);
                    setSource("");
                  })
                }
              />
            </section>
          )}
          {spec.questions.length > 0 && (
            <section className="card paper-question-review">
              <div className="section-heading">
                <h2>Edit paper & marking scheme</h2>
                <span>{spec.questions.length} questions</span>
              </div>
              <MarkingRules
                numbers={spec.questions.map((_, i) => i + 1)}
                disabled={busy}
                onApply={async (numbers, marks, penalty) => {
                  await run(async () => {
                    await save({
                      ...spec,
                      questions: spec.questions.map((q, i) =>
                        numbers.includes(i + 1) ? { ...q, marks, penalty } : q,
                      ),
                    });
                    setReviewed(false);
                    setNotice("Marking scheme saved.");
                  });
                }}
              />
              {spec.questions.map((q, i) => (
                <article className="draft-question" key={i}>
                  <div className="section-heading">
                    <h3>Question {i + 1}</h3>
                    <span>
                      +{q.marks} / −{q.penalty}
                    </span>
                  </div>
                  <p>{q.stem}</p>
                  <div className="draft-answer-options">
                    {q.options?.map((o) => (
                      <label key={o.id} className="check">
                        <input
                          type={q.type === "multiple" ? "checkbox" : "radio"}
                          name={`draft-answer-${i}`}
                          checked={q.key?.includes(o.id) ?? false}
                          onChange={(e) =>
                            change({
                              ...spec,
                              questions: spec.questions.map((old, j) =>
                                j === i
                                  ? {
                                      ...old,
                                      key:
                                        q.type === "multiple"
                                          ? e.target.checked
                                            ? [...(old.key ?? []), o.id]
                                            : (old.key ?? []).filter(
                                                (k) => k !== o.id,
                                              )
                                          : [o.id],
                                    }
                                  : old,
                              ),
                            })
                          }
                        />
                        <span>
                          {o.id}. {o.text}
                        </span>
                      </label>
                    ))}
                  </div>
                  {!q.key?.length &&
                    ["single", "multiple"].includes(q.type) && (
                      <p className="muted">
                        Choose the correct answer or apply a separate answer
                        key.
                      </p>
                    )}
                  <div className="actions">
                    <button
                      disabled={busy}
                      onClick={() => {
                        setEdit({ index: i, question: q });
                        setSource("");
                      }}
                    >
                      Edit question {i + 1}
                    </button>
                    <button
                      disabled={busy}
                      onClick={() =>
                        change({
                          ...spec,
                          questions: spec.questions.filter((_, j) => j !== i),
                        })
                      }
                    >
                      Remove question {i + 1}
                    </button>
                    <button
                      disabled={busy || i === 0}
                      onClick={() => {
                        const qs = [...spec.questions];
                        [qs[i - 1], qs[i]] = [qs[i], qs[i - 1]];
                        change({ ...spec, questions: qs });
                      }}
                    >
                      Move up
                    </button>
                  </div>
                </article>
              ))}
            </section>
          )}
          {preview && preview.state === "preview" && (
            <section className="card">
              <KeyImportPanel
                simple
                preview={preview}
                api={api}
                jobs={[]}
                onApplied={(p) => {
                  setPreview(p);
                  change({
                    ...spec,
                    questions: spec.questions.map((q) => {
                      const row = p.rows.find(
                        (r) => q.id === `source:${p.id}:${r.index}`,
                      );
                      return row ? { ...q, key: row.question.key } : q;
                    }),
                  });
                }}
              />
            </section>
          )}
          {(spec.questions.length > 0 ||
            spec.source_applied ||
            spec.applied_jobs?.length > 0) && (
            <div className="paper-edit-footer">
              <div className="actions">
                <button
                  disabled={processing}
                  onClick={() => setSource("manual")}
                >
                  Add manual question
                </button>
                <button disabled={processing} onClick={() => setSource("ai")}>
                  Add AI questions
                </button>
              </div>
              <label className="check">
                <input
                  type="checkbox"
                  checked={reviewed}
                  onChange={(e) => setReviewed(e.target.checked)}
                />
                I reviewed every question, answer, solution and marking rule.
              </label>
              <button
                className="primary"
                disabled={
                  busy || processing || !reviewed || !spec.questions.length
                }
                onClick={() =>
                  void run(async () => {
                    const saved = await save();
                    const published = await api.publishDraft(
                      saved.id,
                      saved.version,
                    );
                    setDraft(published);
                    setNotice(
                      spec.learners.length
                        ? "Assessment published and assigned."
                        : "Assessment published. Ready to check answer sheets.",
                    );
                  })
                }
              >
                {spec.learners.length
                  ? "Publish & assign assessment"
                  : "Publish assessment"}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export function FacultyStudents({
  api,
  courses,
}: {
  api: EducationClient;
  courses: A<"AcademicCourse">[];
}) {
  const [course, setCourse] = useState(""),
    [rows, setRows] = useState<A<"AcademicRecipient">[]>([]),
    [query, setQuery] = useState(""),
    [error, setError] = useState(""),
    [loading, setLoading] = useState(false);
  useEffect(() => {
    let alive = true;
    setRows([]);
    setError("");
    if (course) {
      setLoading(true);
      void api
        .courseLearners(course)
        .then((r) => {
          if (alive) setRows(r);
        })
        .catch((e) => {
          if (alive) setError(e.message);
        })
        .finally(() => {
          if (alive) setLoading(false);
        });
    }
    return () => {
      alive = false;
    };
  }, [api, course]);
  return (
    <section className="card">
      <h2>Your students</h2>
      <p className="muted">
        Only batches covered by your teaching responsibilities are available.
      </p>
      <div className="academic-columns">
        <Field name="Student batch">
          <select value={course} onChange={(e) => setCourse(e.target.value)}>
            <option value="">Choose a batch</option>
            {courses.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name} · {c.academic_year}
              </option>
            ))}
          </select>
        </Field>
        <Field name="Search students">
          <input
            type="search"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </Field>
      </div>
      {error && (
        <p role="alert" className="error">
          {error}
        </p>
      )}
      {loading ? (
        <p role="status">Loading students…</p>
      ) : (
        <div
          role="region"
          aria-label="Students in your batch"
          className="results-table-wrap"
        >
          <table className="results-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Student reference</th>
              </tr>
            </thead>
            <tbody>
              {rows
                .filter((r) =>
                  `${r.name} ${r.reference}`
                    .toLowerCase()
                    .includes(query.toLowerCase()),
                )
                .map((r) => (
                  <tr key={r.id}>
                    <th scope="row">{r.name}</th>
                    <td>{r.reference}</td>
                  </tr>
                ))}
            </tbody>
          </table>
          {!rows.length && (
            <p className="empty-message">
              {course
                ? "No active students in this batch."
                : "Choose a batch to view your students."}
            </p>
          )}
        </div>
      )}
    </section>
  );
}
