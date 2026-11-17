"use client";
import { useEffect, useState } from "react";
import { type A, type EducationClient } from "../../../clients/academic";
import { FormField as Field } from "./form-field";
import { type LearningView } from "./learning-navigation";
type Paper = A<"AcademicPaperRecord">;
type Job = A<"AcademicJob">;
type Topic = A<"AcademicTopic">;
type Course = A<"AcademicCourse">;
export function LearningHome({
  navigate,
  papers,
  topics,
  jobs,
  openPaper,
}: {
  navigate: (view: LearningView) => void;
  openPaper: (paper: Paper) => void;
  papers: Paper[];
  topics: Topic[];
  jobs: Job[];
}) {
  return (
    <>
      <section className="learning-welcome">
        <p className="eyebrow">Teaching workspace</p>
        <h2>What would you like to do?</h2>
        <p>
          Start with questions, turn them into a test, then collect and review
          answers.
        </p>
        <div className="actions">
          <button className="primary" onClick={() => navigate("generate")}>
            Create questions with AI
          </button>
          <button onClick={() => navigate("import")}>
            Upload a question paper
          </button>
        </div>
      </section>
      <div className="learning-home-grid">
        <section className="card">
          <h2>A clear path from questions to results</h2>
          <ol className="task-path">
            {[
              [
                "curriculum",
                "Prepare questions",
                "Write or import questions and check their answer keys.",
              ],
              [
                "papers",
                "Create an assessment",
                "Select approved questions and assign the test to learners.",
              ],
              [
                "scan",
                "Collect paper answers",
                "Print matching sheets, upload scans and check uncertain marks.",
              ],
              [
                "results",
                "Review the results",
                "Approve written marks and see exactly how scores were awarded.",
              ],
            ].map(([view, title, description], i) => (
              <li key={view}>
                <span className="step-number">{i + 1}</span>
                <div>
                  <button
                    className="text-action"
                    onClick={() => navigate(view as LearningView)}
                  >
                    {title}
                  </button>
                  <p>{description}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>
        <section className="card sample-invitation">
          <p className="eyebrow">First time here?</p>
          <h2>Try it with a ready-made example.</h2>
          <p>
            Create a practice test for a test learner. Download the paper,
            answer key and completed OMR sheets, then upload one to see a real
            result.
          </p>
          <button onClick={() => navigate("samples")}>Try sample files</button>
          <div className="workspace-facts">
            <span>{topics.length} topics</span>
            <span>{papers.length} assessments</span>
            <button className="text-action" onClick={() => navigate("jobs")}>
              {
                jobs.filter((j) => ["queued", "running"].includes(j.state))
                  .length
              }{" "}
              tasks processing
            </button>
          </div>
        </section>
      </div>
      {papers.length > 0 && (
        <section className="card">
          <div className="section-heading">
            <h2>Recent assessments</h2>
            <button onClick={() => navigate("papers")}>See assessments</button>
          </div>
          {papers.slice(0, 3).map((p) => (
            <div className="academic-row" key={p.id}>
              <div>
                <strong>{p.spec.title}</strong>
                <p className="muted">
                  {p.reference} ·{" "}
                  {p.state === "draft"
                    ? "Draft — needs publishing"
                    : "Published"}{" "}
                  · {p.maximum} marks
                </p>
              </div>
              <button onClick={() => openPaper(p)}>Open assessment</button>
            </div>
          ))}
        </section>
      )}
    </>
  );
}
export function GenerationFlow({
  api,
  topics,
  courses,
  onTopic,
  onReview,
  initialTopic,
}: {
  initialTopic: string;
  api: EducationClient;
  topics: Topic[];
  courses: Course[];
  onTopic: (id: string) => void;
  onReview: (id: string) => void;
}) {
  const [selection, setSelection] = useState(initialTopic);
  const [createdTopic, setCreatedTopic] = useState<Topic | null>(null);
  const availableTopics =
    createdTopic && !topics.some((t) => t.id === createdTopic.id)
      ? [...topics, createdTopic]
      : topics;
  const [job, setJob] = useState<Job | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const chosen = selection || availableTopics[0]?.id || "new";
  useEffect(() => {
    if (!job || !["queued", "running"].includes(job.state)) return;
    let active = true;
    let pending = false;
    const poll = async () => {
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
              : "Connection lost. Your request is saved; refresh its status.",
          );
      } finally {
        pending = false;
      }
    };
    const timer = setInterval(() => void poll(), 3000);
    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [api, job?.id, job?.state]);
  const running = !!job && ["queued", "running"].includes(job.state);
  const labels: Record<string, string> = {
    queued: "Waiting to start",
    running: "Creating draft questions",
    completed: "Your draft questions are ready",
    failed: "The questions could not be generated",
    configuration_required: "The AI provider needs attention",
    cancelled: "Request cancelled",
  };
  async function submit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const f = new FormData(e.currentTarget);
    try {
      let id = chosen;
      if (chosen === "new") {
        const course = courses.find((c) => c.id === String(f.get("course")));
        if (!course)
          throw new Error(
            "Choose a course first. A coordinator can add one under People & organization.",
          );
        const t = await api.topics({
          id: "",
          campus_id: course.campus_id,
          course_id: course.id,
          subject: String(f.get("subject")),
          name: String(f.get("name")),
          outcome: String(f.get("outcome")),
          language: "en",
        });
        id = t.id;
        setCreatedTopic(t);
        setSelection(id);
      }
      onTopic(id);
      setJob(
        await api.generate({
          topic_id: id,
          count: Number(f.get("count")),
          difficulty: String(f.get("difficulty")),
        }),
      );
    } catch (e) {
      setError(
        e instanceof Error
          ? e.message
          : "Could not start generation. Your input is still here.",
      );
    } finally {
      setBusy(false);
    }
  }
  return (
    <div className="generation-layout">
      <section className="card">
        <h2>Tell us what to teach</h2>
        <p>
          Choose a topic or describe a new one. Generated questions are saved as
          drafts for your review.
        </p>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        <form onSubmit={submit}>
          <fieldset disabled={busy || running}>
            <legend className="sr-only">Question generation</legend>
            <Field name="Topic to generate questions for">
              <select
                value={chosen}
                onChange={(e) => setSelection(e.target.value)}
              >
                {availableTopics.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.subject} / {t.name}
                  </option>
                ))}
                <option value="new">+ Describe a new topic</option>
              </select>
            </Field>
            {chosen === "new" ? (
              <>
                <Field name="Course">
                  <select name="course" required>
                    <option value="">Choose a course</option>
                    {courses.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field name="Subject">
                  <input name="subject" required defaultValue="Mathematics" />
                </Field>
                <Field name="Topic name">
                  <input
                    name="name"
                    required
                    minLength={2}
                    placeholder="For example, adding fractions"
                  />
                </Field>
                <Field name="What should learners be able to do?">
                  <textarea
                    name="outcome"
                    required
                    minLength={5}
                    placeholder="For example, add fractions with the same denominator and explain each step."
                  />
                </Field>
              </>
            ) : (
              <p className="topic-context">
                {availableTopics.find((t) => t.id === chosen)?.outcome}
              </p>
            )}
            <div className="academic-columns">
              <Field name="How many questions?">
                <input
                  name="count"
                  type="number"
                  min={1}
                  max={20}
                  defaultValue={5}
                  required
                />
              </Field>
              <Field name="Difficulty">
                <select name="difficulty">
                  <option value="introductory">Easy — introductory</option>
                  <option value="standard">Standard</option>
                  <option value="challenging">Challenging</option>
                </select>
              </Field>
            </div>
            <button className="primary" disabled={!courses.length}>
              {busy ? "Starting…" : "Generate draft questions"}
            </button>
          </fieldset>
        </form>
      </section>
      <section className="card generation-status" aria-live="polite">
        <p className="eyebrow">Request status</p>
        <h2>
          {job
            ? (labels[job.state] ?? job.state)
            : "Your questions will appear here"}
        </h2>
        {!job ? (
          <>
            <p>
              No request has been sent yet. Your institution's configured AI
              provider will receive the topic and learning outcome.
            </p>
            <ol>
              <li>Describe the learning goal.</li>
              <li>Wait for the saved request to finish.</li>
              <li>Check each question, answer and explanation.</li>
            </ol>
          </>
        ) : (
          <>
            <p>
              {running
                ? "You can leave this screen. The request is saved and continues in the background. Status updates every few seconds."
                : job.state === "completed"
                  ? "Review every answer key before approving questions for a test."
                  : "Your topic and request are saved. Existing questions and assessments are unaffected."}
            </p>
            {job.error && <p className="error">{job.error}</p>}
            <p>
              <a href="#education/samples">Try ready-made sample questions</a> ·{" "}
              <a href="#education/curriculum">Write questions yourself</a>
            </p>
            <p className="muted">
              {job.provider
                ? `Provider: ${job.provider} ${job.model}`
                : "Provider details appear after a response is received."}
            </p>
            <div className="actions">
              {job.state === "completed" && (
                <button
                  className="primary"
                  onClick={() => onReview(job.resource_id)}
                >
                  Review generated questions
                </button>
              )}
              {["failed", "configuration_required", "cancelled"].includes(
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
                  Retry this request
                </button>
              )}
              {running && (
                <button
                  disabled={busy}
                  onClick={async () => {
                    try {
                      await api.jobAction(job.id, "cancel");
                      setJob(await api.detail("jobs", job.id));
                    } catch (e) {
                      setError(
                        e instanceof Error ? e.message : "Could not cancel",
                      );
                    }
                  }}
                >
                  Cancel request
                </button>
              )}
              <button
                onClick={() =>
                  void api
                    .detail("jobs", job.id)
                    .then(setJob)
                    .catch((e) => setError(e.message))
                }
              >
                Check status now
              </button>
            </div>
            {job.provider === "development_fixture" && (
              <p className="environment">
                Demonstration output. This was not generated by an external AI
                service.
              </p>
            )}
          </>
        )}
      </section>
    </div>
  );
}
export function SamplesFlow({
  api,
  courses,
  onOpen,
}: {
  api: EducationClient;
  courses: Course[];
  onOpen: (kit: A<"AcademicSampleKit">) => void;
}) {
  const [course, setCourse] = useState("");
  const [learners, setLearners] = useState<A<"AcademicRecipient">[]>([]);
  const [learner, setLearner] = useState("");
  const [kit, setKit] = useState<A<"AcademicSampleKit"> | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    setLearner("");
    setLearners([]);
    if (course)
      api
        .courseLearners(course)
        .then((l) => {
          if (active) setLearners(l);
        })
        .catch((e) => {
          if (active) setError(e.message);
        });
    return () => {
      active = false;
    };
  }, [api, course]);
  return (
    <>
      <section className="sample-intro">
        <p className="eyebrow">Original, ready-to-use test material</p>
        <h2>Try the full flow in a few minutes.</h2>
        <p>
          Create a practice test for a test learner. We will give you matching
          files, including filled answer sheets with known scores. No AI
          provider is needed.
        </p>
        <div className="sample-downloads">
          <a
            href="/samples/question-paper.pdf"
            target="_blank"
            rel="noreferrer"
          >
            Preview the question paper ↗
          </a>
          <a href="/samples/marking-guide.pdf" target="_blank" rel="noreferrer">
            See the answer key ↗
          </a>
          <a href="/samples/questions.csv" download>
            Download questions CSV
          </a>
          <a href="/samples/questions.json" download>
            Download questions JSON
          </a>
        </div>
      </section>
      <div className="academic-columns">
        <section className="card">
          <h2>1. Prepare your sample test</h2>
          <p>
            This creates a new six-question practice assessment and assigns it
            to the learner below. Choose a test account.
          </p>
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              setBusy(true);
              setError("");
              try {
                setKit(
                  await api.sampleKit({
                    course_id: course,
                    learner_id: learner,
                  }),
                );
              } catch (e) {
                setError(
                  e instanceof Error
                    ? e.message
                    : "Could not prepare samples. Please retry.",
                );
              } finally {
                setBusy(false);
              }
            }}
          >
            <Field name="Sample course">
              <select
                value={course}
                required
                onChange={(e) => {
                  setCourse(e.target.value);
                  setKit(null);
                }}
              >
                <option value="">Choose a course</option>
                {courses.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </Field>
            <Field name="Test learner">
              <select
                value={learner}
                required
                disabled={!course}
                onChange={(e) => setLearner(e.target.value)}
              >
                <option value="">Choose a test learner</option>
                {learners.map((l) => (
                  <option key={l.id} value={l.id}>
                    {l.name} / {l.reference}
                  </option>
                ))}
              </select>
            </Field>
            {course && learners.length === 0 && (
              <p className="muted">
                This course has no eligible active learners. Add an enrollment
                under People & organization, or choose the manual test course.
              </p>
            )}
            <button className="primary" disabled={busy || !learner}>
              {busy ? "Preparing files…" : "Create a sample practice test"}
            </button>
          </form>
          {kit && (
            <div className="notice" role="status">
              <strong>Your matching files are ready.</strong>
              <p>The sheet identity belongs to this new practice test.</p>
              <a
                className="button-link"
                href={api.url(
                  `sample-files/${kit.paper_id}?assignment=${kit.assignment_id}`,
                )}
                download
              >
                Download complete test kit (.zip)
              </a>
              <button onClick={() => onOpen(kit)}>Go to Scan & check</button>
            </div>
          )}
        </section>
        <section className="card">
          <h2>2. Upload one filled sheet</h2>
          <ol className="sample-instructions">
            <li>
              <strong>Download and unzip the kit.</strong>
              <p>
                Start with <code>omr-correct.png</code>. The PDF version tests
                scanned-PDF intake.
              </p>
            </li>
            <li>
              <strong>Open Scan & check.</strong>
              <p>
                Choose the new sample test, upload the sheet and wait for
                processing.
              </p>
            </li>
            <li>
              <strong>Check identity and save the result.</strong>
              <p>
                The correctly filled sheet scores 12 / 12. The mixed-answer
                sheet scores 5 / 12.
              </p>
            </li>
          </ol>
          <details>
            <summary>What else can I test?</summary>
            <p>
              <code>omr-needs-review.png</code> has two marks on question 1:
              correct it to B before committing.{" "}
              <code>duplicate-pages.pdf</code> repeats the same page to test
              deduplication. Print <code>omr-blank.pdf</code> to test a real
              scanner or camera.
            </p>
            <p>
              Create a fresh sample test before committing a different example:
              each learner assignment keeps one sitting.
            </p>
          </details>
          <p className="muted">
            Original CC0-1.0 demonstration content. These registered sheets are
            compatible with this reader; unrelated commercial OMR layouts are
            not automatically supported.
          </p>
        </section>
      </div>
    </>
  );
}
export function ResultsFlow({
  api,
  papers,
  onSitting,
  initialPaper,
}: {
  initialPaper: string;
  api: EducationClient;
  papers: Paper[];
  onSitting: (s: A<"AcademicSitting">) => void;
}) {
  const [paper, setPaper] = useState(initialPaper);
  useEffect(() => setPaper(initialPaper), [initialPaper]);
  const [rows, setRows] = useState<A<"AcademicAssigned">[]>([]);
  const [error, setError] = useState("");
  useEffect(() => {
    let active = true;
    setRows([]);
    if (paper)
      api
        .list("assignments", { paper })
        .then((r) => {
          if (active) setRows(r);
        })
        .catch((e) => {
          if (active) setError(e.message);
        });
    return () => {
      active = false;
    };
  }, [api, paper]);
  return (
    <section className="card">
      <Field name="Assessment to review">
        <select value={paper} onChange={(e) => setPaper(e.target.value)}>
          <option value="">Choose an assessment</option>
          {papers
            .filter((p) => p.state === "published")
            .map((p) => (
              <option key={p.id} value={p.id}>
                {p.spec.title}
              </option>
            ))}
        </select>
      </Field>
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {!paper ? (
        <p className="empty-message">
          Choose an assessment to see who has submitted and what needs review.
        </p>
      ) : rows.length === 0 ? (
        <p>
          No learners assigned to this assessment yet. Assign it from
          Assessments.
        </p>
      ) : (
        rows.map((r) => (
          <article className="academic-row" key={r.id}>
            <div>
              <strong>{r.learner}</strong>
              <p className="muted">{r.state.replaceAll("_", " ")}</p>
            </div>
            {r.sitting_id ? (
              <button
                onClick={() =>
                  void api
                    .detail("sittings", r.sitting_id)
                    .then(onSitting)
                    .catch((e) => setError(e.message))
                }
              >
                {r.state === "awaiting_review"
                  ? "Mark written answers"
                  : "Open result & answers"}
              </button>
            ) : (
              <span className="muted">Not started</span>
            )}
          </article>
        ))
      )}
    </section>
  );
}
