"use client";
import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type PointerEvent,
  type KeyboardEvent,
} from "react";
import {
  autoCount,
  blockCircles,
  loadProfile,
  nativeFiles,
  referenceFile,
  type AnswerBlock,
  type DesignerLayout,
  type IdentityKind,
  type OMRProfile,
  type Point,
  type Rect,
} from "../../../clients/omr-template";
import { FormField as Field } from "./form-field";

const empty = (): DesignerLayout => ({
  width: 0,
  height: 0,
  name: "OMR layout",
  identity: {},
  blocks: [],
});
function download(name: string, data: BlobPart, type = "application/json") {
  const url = URL.createObjectURL(new Blob([data], { type }));
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}
async function encoded(blob: Blob) {
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(String(reader.result).split(",")[1]);
    reader.onerror = reject;
    reader.readAsDataURL(blob);
  });
}
export function OMRTemplateBuilder({
  file,
  onProfile,
}: {
  file: File | null;
  onProfile: (profile: OMRProfile | null) => void;
}) {
  const [source, setSource] = useState<File | null>(null),
    [url, setUrl] = useState("");
  const [layout, setLayout] = useState<DesignerLayout>(empty),
    [step, setStep] = useState(1);
  const [tool, setTool] = useState<IdentityKind | "block" | "points" | "none">(
    "student_name",
  );
  const [draft, setDraft] = useState<AnswerBlock | null>(null),
    [draw, setDraw] = useState<Rect | null>(null);
  const [profile, setProfile] = useState<OMRProfile | null>(null),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const [zoom, setZoom] = useState(100),
    [options, setOptions] = useState(4);
  const [cursor, setCursor] = useState<Point | null>(null);
  const keyboardStart = useRef<Point | null>(null);
  const keyboardHelp = useId();
  const applied = useRef(false),
    image = useRef<HTMLImageElement | null>(null),
    svg = useRef<SVGSVGElement | null>(null);
  const drag = useRef<{ start: Point; rect?: Rect; pointer: number } | null>(
    null,
  );
  useEffect(() => {
    if (file && !applied.current) {
      setSource(file);
      setLayout(empty());
      setStep(1);
      setTool("student_name");
      setDraft(null);
      setError("");
    }
  }, [file]);
  useEffect(() => {
    if (!source) return;
    const next = URL.createObjectURL(source);
    setUrl(next);
    return () => URL.revokeObjectURL(next);
  }, [source]);
  const draftInfo = useMemo(() => {
    if (!draft || draft.points.length !== draft.options * 2)
      return { rows: [], error: "" };
    try {
      return {
        rows: blockCircles(draft, layout.width, layout.height),
        error: "",
      };
    } catch (e) {
      return {
        rows: [],
        error: e instanceof Error ? e.message : "Check the clicked circles.",
      };
    }
  }, [draft, layout.width, layout.height]);
  const rows = useMemo(
    () =>
      layout.blocks.flatMap((b) =>
        blockCircles(b, layout.width, layout.height),
      ),
    [layout.blocks, layout.width, layout.height],
  );
  const total = layout.blocks.reduce((sum, b) => sum + b.count, 0);
  function invalidate() {
    applied.current = false;
    setProfile(null);
    onProfile(null);
    setError("");
  }
  function point(e: PointerEvent<SVGSVGElement>): Point {
    const r = e.currentTarget.getBoundingClientRect();
    return {
      x: Math.max(0, Math.min(1, (e.clientX - r.left) / r.width)),
      y: Math.max(0, Math.min(1, (e.clientY - r.top) / r.height)),
    };
  }
  function down(e: PointerEvent<SVGSVGElement>) {
    if (e.button !== 0 || tool === "none" || profile || !layout.width) return;
    e.preventDefault();
    e.currentTarget.focus();
    const p = point(e);
    if (tool === "points") {
      addPoint(p);
      return;
    }
    drag.current = { start: p, pointer: e.pointerId };
    setDraw({ x: p.x, y: p.y, width: 0, height: 0 });
    e.currentTarget.setPointerCapture(e.pointerId);
  }
  function addPoint(p: Point) {
    if (!draft || draft.points.length >= draft.options * 2) return;
    const next = { ...draft, points: [...draft.points, p] };
    if (next.points.length === next.options * 2) {
      try {
        next.count = autoCount(next, layout.width, layout.height);
      } catch {
        /* Keep clicks available for undo. */
      }
    }
    setDraft(next);
    setError("");
  }
  function move(e: PointerEvent<SVGSVGElement>) {
    const g = drag.current;
    if (!g) return;
    const p = point(e);
    g.rect = {
      x: Math.min(g.start.x, p.x),
      y: Math.min(g.start.y, p.y),
      width: Math.abs(p.x - g.start.x),
      height: Math.abs(p.y - g.start.y),
    };
    setDraw(g.rect);
  }
  function up(e: PointerEvent<SVGSVGElement>) {
    const g = drag.current;
    if (!g) return;
    drag.current = null;
    setDraw(null);
    if (e.currentTarget.hasPointerCapture(g.pointer))
      e.currentTarget.releasePointerCapture(g.pointer);
    finishBox(g.rect);
  }
  function finishBox(r?: Rect) {
    if (!r || r.width * layout.width < 5 || r.height * layout.height < 5) {
      setError("Draw a larger box around the field or answer block.");
      return;
    }
    if (tool === "student_name" || tool === "roll_number") {
      setLayout({ ...layout, identity: { ...layout.identity, [tool]: r } });
      setTool(tool === "student_name" ? "roll_number" : "none");
    } else if (tool === "block") {
      const first = Math.max(1, ...layout.blocks.map((b) => b.first + b.count));
      setDraft({ bounds: r, points: [], first, count: 1, options, size: 0.38 });
      setTool("points");
    }
    setError("");
  }
  function cancel() {
    keyboardStart.current = null;
    const g = drag.current;
    drag.current = null;
    setDraw(null);
    if (g && svg.current?.hasPointerCapture(g.pointer))
      svg.current.releasePointerCapture(g.pointer);
  }
  function keyboard(e: KeyboardEvent<SVGSVGElement>) {
    if (e.key === "Escape") {
      e.preventDefault();
      cancel();
      return;
    }
    if (tool === "none" || !layout.width || !layout.height) return;
    const p = cursor ?? { x: 0.5, y: 0.5 };
    if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(e.key)) {
      e.preventDefault();
      const pixels = e.shiftKey ? 50 : e.altKey ? 1 : 5;
      const next = {
        x: Math.max(
          0,
          Math.min(
            1,
            p.x +
              (e.key === "ArrowRight"
                ? pixels
                : e.key === "ArrowLeft"
                  ? -pixels
                  : 0) /
                layout.width,
          ),
        ),
        y: Math.max(
          0,
          Math.min(
            1,
            p.y +
              (e.key === "ArrowDown"
                ? pixels
                : e.key === "ArrowUp"
                  ? -pixels
                  : 0) /
                layout.height,
          ),
        ),
      };
      setCursor(next);
      const a = keyboardStart.current;
      if (a)
        setDraw({
          x: Math.min(a.x, next.x),
          y: Math.min(a.y, next.y),
          width: Math.abs(a.x - next.x),
          height: Math.abs(a.y - next.y),
        });
    } else if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      setCursor(p);
      if (tool === "points") addPoint(p);
      else if (keyboardStart.current) {
        const a = keyboardStart.current;
        keyboardStart.current = null;
        setDraw(null);
        finishBox({
          x: Math.min(a.x, p.x),
          y: Math.min(a.y, p.y),
          width: Math.abs(a.x - p.x),
          height: Math.abs(a.y - p.y),
        });
      } else keyboardStart.current = p;
    }
  }
  function addBlock() {
    if (!draft) return;
    try {
      const next = { ...layout, blocks: [...layout.blocks, draft] };
      nativeFiles(next);
      setLayout(next);
      setDraft(null);
      setTool("block");
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Check this block.");
    }
  }
  async function create() {
    setBusy(true);
    setError("");
    try {
      nativeFiles(layout);
      const img = image.current;
      if (!img?.complete) throw Error("Wait for the sheet image to load.");
      const canvas = document.createElement("canvas");
      canvas.width = layout.width;
      canvas.height = layout.height;
      const ctx = canvas.getContext("2d")!;
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
      ctx.fillStyle = "#fff";
      for (const r of Object.values(layout.identity))
        ctx.fillRect(
          r.x * canvas.width,
          r.y * canvas.height,
          r.width * canvas.width,
          r.height * canvas.height,
        );
      for (const row of rows)
        for (const p of row.points) {
          ctx.fillStyle = "#fff";
          ctx.beginPath();
          ctx.arc(
            p.x * canvas.width,
            p.y * canvas.height,
            row.size * 0.6,
            0,
            Math.PI * 2,
          );
          ctx.fill();
          ctx.strokeStyle = "#777";
          ctx.lineWidth = 1.5;
          ctx.beginPath();
          ctx.arc(
            p.x * canvas.width,
            p.y * canvas.height,
            row.size * 0.48,
            0,
            Math.PI * 2,
          );
          ctx.stroke();
        }
      const toBlob = (type: string) =>
        new Promise<Blob>((resolve, reject) =>
          canvas.toBlob(
            (b) =>
              b
                ? resolve(b)
                : reject(Error("Could not create the reference sheet.")),
            type,
            0.96,
          ),
        );
      let blob = await toBlob("image/png");
      if (blob.size > 8 * 1024 * 1024) blob = await toBlob("image/jpeg");
      if (blob.size > 8 * 1024 * 1024)
        throw Error("Use a smaller image for this layout.");
      const name =
        blob.type === "image/png" ? "reference.png" : "reference.jpg";
      const next: OMRProfile = {
        version: 1,
        layout,
        ...nativeFiles(layout, name),
        reference: { name, type: blob.type, data: await encoded(blob) },
      };
      // Keep generated native files tied to the same geometry and reference.
      applied.current = true;
      setProfile(next);
      onProfile(next);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create the layout.");
    } finally {
      setBusy(false);
    }
  }
  const instructions =
    tool === "student_name"
      ? "Draw a box around the student name."
      : tool === "roll_number"
        ? "Draw a box around the roll number."
        : tool === "points" && draft
          ? draft.points.length < draft.options * 2
            ? `Row ${Math.floor(draft.points.length / draft.options) + 1}: click the center of ${String.fromCharCode(65 + (draft.points.length % draft.options))}. ${draft.points.length} of ${draft.options * 2} circles selected.`
            : "The remaining circles are filled in. Check the row count, then add this block."
          : tool === "block"
            ? "Draw a box around one column of answer bubbles."
            : "Name and roll fields are set. Continue to the answers.";
  return (
    <section className="omr-designer" aria-label="Guided OMR layout setup">
      <div className="section-heading">
        <div>
          <h3>{profile ? "Your layout is ready" : "Create your OMR layout"}</h3>
          <p className="muted">
            Use a blank sheet when available. The template, reference image and
            tuning are created together.
          </p>
        </div>
        <label className="omr-load-calibration">
          Load saved layout
          <input
            type="file"
            accept="application/json,.json"
            aria-label="Load saved OMR layout"
            disabled={busy}
            onChange={async (e) => {
              const f = e.target.files?.[0];
              if (!f) return;
              try {
                if (f.size > 12_000_000)
                  throw Error("Saved layout is too large.");
                const next = loadProfile(await f.text());
                setLayout(next.layout);
                setSource(referenceFile(next));
                setStep(3);
                setDraft(null);
                applied.current = true;
                setProfile(next);
                onProfile(next);
                setError("");
              } catch (err) {
                setError(
                  err instanceof Error
                    ? err.message
                    : "Could not load this layout.",
                );
              }
              e.target.value = "";
            }}
          />
        </label>
      </div>
      {profile ? (
        <div className="omr-layout-ready">
          <strong>
            {layout.name} · {total} questions · {layout.blocks.length} answer
            blocks
          </strong>
          <p>
            Reference alignment is configured automatically. No marker upload is
            needed.
          </p>
          <div className="actions">
            <button
              type="button"
              onClick={() =>
                download(
                  `${layout.name.replace(/[^a-z0-9]+/gi, "-")}.omr.json`,
                  JSON.stringify(profile),
                )
              }
            >
              Save reusable layout
            </button>
            <button
              type="button"
              onClick={() => {
                invalidate();
                setStep(2);
                setTool("block");
              }}
            >
              Edit layout
            </button>
            <button
              type="button"
              onClick={() => {
                invalidate();
                setSource(file);
                setLayout(empty());
                setStep(1);
                setTool("student_name");
                setDraft(null);
              }}
            >
              Start a new layout
            </button>
          </div>
          <details>
            <summary>Generated OMRChecker files</summary>
            <div className="actions">
              <button
                type="button"
                onClick={() =>
                  download(
                    "template.json",
                    JSON.stringify(profile.template, null, 2),
                  )
                }
              >
                Download template
              </button>
              <button
                type="button"
                onClick={() =>
                  download(
                    "config.json",
                    JSON.stringify(profile.config, null, 2),
                  )
                }
              >
                Download tuning
              </button>
              <button
                type="button"
                onClick={() => {
                  const f = referenceFile(profile);
                  download(f.name, f, f.type);
                }}
              >
                Download reference image
              </button>
            </div>
          </details>
        </div>
      ) : (
        <>
          <ol className="omr-setup-steps">
            {["Name & roll", "Answer blocks", "Finish"].map((name, i) => (
              <li key={name} aria-current={step === i + 1 ? "step" : undefined}>
                {i + 1}. {name}
              </li>
            ))}
          </ol>
          {!source ? (
            <p>Choose a sheet image above, or load a saved layout.</p>
          ) : (
            <>
              {step === 1 && (
                <div className="actions">
                  <button
                    type="button"
                    aria-pressed={tool === "student_name"}
                    onClick={() => setTool("student_name")}
                  >
                    Snip student name{layout.identity.student_name ? " ✓" : ""}
                  </button>
                  <button
                    type="button"
                    aria-pressed={tool === "roll_number"}
                    onClick={() => setTool("roll_number")}
                  >
                    Snip roll number{layout.identity.roll_number ? " ✓" : ""}
                  </button>
                  <button
                    type="button"
                    className="primary"
                    onClick={() => {
                      setStep(2);
                      setTool("block");
                    }}
                  >
                    Continue to answer boxes
                  </button>
                  <button
                    type="button"
                    onClick={() => setLayout({ ...layout, identity: {} })}
                  >
                    Clear identity boxes
                  </button>
                </div>
              )}
              {step === 2 && (
                <>
                  <div className="omr-canvas-controls">
                    <Field name="Options per question">
                      <select
                        value={options}
                        disabled={!!draft}
                        onChange={(e) => setOptions(Number(e.target.value))}
                      >
                        {[2, 3, 4, 5, 6, 7, 8].map((n) => (
                          <option key={n} value={n}>
                            {n}
                          </option>
                        ))}
                      </select>
                    </Field>
                    <button
                      type="button"
                      onClick={() => {
                        setDraft(null);
                        setTool("block");
                        setError("");
                      }}
                    >
                      Draw answer box
                    </button>
                    <button
                      type="button"
                      disabled={!draft?.points.length}
                      onClick={() =>
                        setDraft(
                          draft
                            ? { ...draft, points: draft.points.slice(0, -1) }
                            : null,
                        )
                      }
                    >
                      Undo last circle
                    </button>
                  </div>
                  {draft && (
                    <div className="omr-grid-setup">
                      <Field name="Block first question">
                        <input
                          type="number"
                          min={1}
                          value={draft.first}
                          onChange={(e) =>
                            setDraft({
                              ...draft,
                              first: Number(e.target.value),
                            })
                          }
                        />
                      </Field>
                      <Field name="Questions in this box">
                        <input
                          type="number"
                          min={1}
                          max={200}
                          value={draft.count}
                          onChange={(e) =>
                            setDraft({
                              ...draft,
                              count: Number(e.target.value),
                            })
                          }
                        />
                      </Field>
                      {draft.points.length === draft.options * 2 && (
                        <button
                          type="button"
                          className="primary"
                          disabled={!!draftInfo.error}
                          onClick={addBlock}
                        >
                          Add answer block
                        </button>
                      )}
                      <details>
                        <summary>Bubble size</summary>
                        <input
                          aria-label="Bubble sample size"
                          type="range"
                          min={0.15}
                          max={0.8}
                          step={0.01}
                          value={draft.size}
                          onChange={(e) =>
                            setDraft({ ...draft, size: Number(e.target.value) })
                          }
                        />
                      </details>
                    </div>
                  )}
                  {!!layout.blocks.length && (
                    <div className="omr-saved-blocks">
                      {layout.blocks.map((b, i) => (
                        <div key={i}>
                          <span>
                            Block {i + 1} · questions {b.first}–
                            {b.first + b.count - 1}
                          </span>
                          <button
                            type="button"
                            onClick={() => {
                              setDraft(b);
                              setLayout({
                                ...layout,
                                blocks: layout.blocks.filter((_, j) => j !== i),
                              });
                              setTool("points");
                            }}
                          >
                            Edit block {i + 1}
                          </button>
                          <button
                            type="button"
                            onClick={() =>
                              setLayout({
                                ...layout,
                                blocks: layout.blocks.filter((_, j) => j !== i),
                              })
                            }
                          >
                            Remove block {i + 1}
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                  <div className="actions">
                    <button
                      type="button"
                      onClick={() => {
                        setStep(1);
                        setTool("student_name");
                      }}
                    >
                      Back to identity fields
                    </button>
                    <button
                      type="button"
                      className="primary"
                      disabled={!layout.blocks.length || !!draft}
                      onClick={() => {
                        setStep(3);
                        setTool("none");
                      }}
                    >
                      Review layout
                    </button>
                  </div>
                </>
              )}
              {step === 3 && (
                <div className="omr-grid-setup">
                  <Field name="OMR layout name">
                    <input
                      value={layout.name}
                      maxLength={80}
                      onChange={(e) =>
                        setLayout({ ...layout, name: e.target.value })
                      }
                    />
                  </Field>
                  <p>
                    {total} questions · {layout.blocks.length} blocks ·{" "}
                    {Object.keys(layout.identity).length} identity fields
                  </p>
                  <button
                    type="button"
                    onClick={() => {
                      setStep(2);
                      setTool("block");
                    }}
                  >
                    Back to answer boxes
                  </button>
                  <button
                    type="button"
                    className="primary"
                    disabled={busy || !layout.name.trim()}
                    onClick={() => void create()}
                  >
                    {busy ? "Creating layout…" : "Use this layout"}
                  </button>
                </div>
              )}
              <div className="omr-canvas-controls">
                <Field name="Layout zoom">
                  <select
                    value={zoom}
                    onChange={(e) => setZoom(Number(e.target.value))}
                  >
                    {[50, 75, 100, 125, 150, 200].map((n) => (
                      <option key={n} value={n}>
                        {n}%
                      </option>
                    ))}
                  </select>
                </Field>
                <p role="status">
                  {step === 3
                    ? "Check the generated circles before using this layout."
                    : instructions}
                </p>
              </div>
              <div className="omr-canvas-viewport">
                <div className="omr-canvas-image" style={{ width: `${zoom}%` }}>
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    ref={image}
                    src={url || undefined}
                    alt="Sheet for guided OMR layout setup"
                    draggable={false}
                    onError={() =>
                      setError("Choose a readable PNG or JPEG sheet image.")
                    }
                    onLoad={(e) => {
                      const img = e.currentTarget;
                      if (img.naturalWidth * img.naturalHeight > 8_000_000) {
                        setError("Use an image of at most 8 megapixels.");
                        return;
                      }
                      const scale = Math.min(
                        1,
                        5000 / Math.max(img.naturalWidth, img.naturalHeight),
                      );
                      setLayout((old) => ({
                        ...old,
                        width: Math.round(img.naturalWidth * scale),
                        height: Math.round(img.naturalHeight * scale),
                      }));
                    }}
                  />
                  <svg
                    ref={svg}
                    viewBox="0 0 1 1"
                    preserveAspectRatio="none"
                    tabIndex={0}
                    role="group"
                    aria-label="Template calibration canvas"
                    aria-describedby={keyboardHelp}
                    className={
                      tool === "points" ? "omr-point-tool" : "omr-draw-tool"
                    }
                    onPointerDown={down}
                    onPointerMove={move}
                    onPointerUp={up}
                    onPointerCancel={cancel}
                    onKeyDown={keyboard}
                    onFocus={() =>
                      setCursor((old) => old ?? { x: 0.5, y: 0.5 })
                    }
                    onBlur={() => {
                      setCursor(null);
                      cancel();
                    }}
                  >
                    {Object.entries(layout.identity).map(([kind, r]) => (
                      <rect key={kind} {...r} className="omr-identity-frame">
                        <title>
                          {kind === "student_name"
                            ? "Student name"
                            : "Roll number"}
                        </title>
                      </rect>
                    ))}
                    {layout.blocks.map((b, i) => (
                      <rect key={i} {...b.bounds} className="omr-block-frame" />
                    ))}
                    {draft && (
                      <rect {...draft.bounds} className="omr-block-frame" />
                    )}
                    {[...rows, ...draftInfo.rows].flatMap((row) =>
                      row.points.map((p, i) => (
                        <ellipse
                          key={`${row.number}-${i}`}
                          cx={p.x}
                          cy={p.y}
                          rx={row.size / 2 / layout.width}
                          ry={row.size / 2 / layout.height}
                          className="omr-bubble"
                        />
                      )),
                    )}
                    {draft?.points.map((p, i) => (
                      <g key={i}>
                        <ellipse
                          cx={p.x}
                          cy={p.y}
                          rx={0.006}
                          ry={(0.006 * layout.width) / layout.height}
                          className="omr-seed-circle"
                        />
                        <text
                          x={p.x + 0.008}
                          y={p.y}
                          fontSize=".012"
                          fill="#3f6f83"
                        >
                          {Math.floor(i / draft.options) + 1}
                          {String.fromCharCode(65 + (i % draft.options))}
                        </text>
                      </g>
                    ))}
                    {draw && <rect {...draw} className="omr-drawing-frame" />}
                    {cursor && (
                      <g
                        stroke="#3f6f83"
                        strokeWidth=".002"
                        pointerEvents="none"
                      >
                        <path
                          d={`M${cursor.x - 0.012} ${cursor.y}h.024 M${cursor.x} ${cursor.y - 0.009}v.018`}
                        />
                      </g>
                    )}
                  </svg>
                </div>
              </div>
              <p id={keyboardHelp} className="muted">
                Keyboard: focus the sheet, use arrow keys to move, then Enter to
                place each corner or circle. Shift moves faster; Alt moves one
                pixel. Escape cancels a box.
              </p>
            </>
          )}
        </>
      )}
      {(error || draftInfo.error) && (
        <p role="alert" className="error">
          {error || draftInfo.error}
        </p>
      )}
    </section>
  );
}
