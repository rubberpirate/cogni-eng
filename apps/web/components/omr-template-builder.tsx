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
  optionValue,
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
const fieldName = (kind: string) =>
  kind === "student_name" ? "Student name" : "Registration number";
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

// Pointer moves only update this transient SVG rectangle, never the React field tree.
function Sheet({
  url,
  layout,
  draft,
  pending,
  clicking,
  guides,
  onBox,
  onPoint,
  onLoad,
  onError,
  imageRef,
}: {
  url: string;
  layout: DesignerLayout;
  draft: AnswerBlock | null;
  pending: Rect | null;
  clicking: boolean;
  guides: boolean;
  onBox: (r: Rect) => void;
  onPoint: (p: Point) => void;
  onLoad: (width: number, height: number) => void;
  onError: () => void;
  imageRef: React.RefObject<HTMLImageElement | null>;
}) {
  const stage = useRef<HTMLDivElement>(null),
    surface = useRef<SVGSVGElement>(null),
    preview = useRef<SVGRectElement>(null);
  const [available, setAvailable] = useState({ width: 600, height: 650 });
  const [zoom, setZoom] = useState("fit"),
    [cursor, setCursor] = useState<Point | null>(null);
  const help = useId(),
    gesture = useRef<{
      start: Point;
      box: DOMRect;
      rect: Rect;
      pointer: number;
    } | null>(null);
  const keyboardStart = useRef<Point | null>(null),
    animation = useRef<number | null>(null);
  useEffect(() => {
    const node = stage.current;
    if (!node) return;
    const observer = new ResizeObserver(() =>
      setAvailable({ width: node.clientWidth, height: node.clientHeight }),
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, []);
  useEffect(
    () => () => {
      if (animation.current !== null) cancelAnimationFrame(animation.current);
    },
    [],
  );
  const aspect =
    layout.width && layout.height ? layout.width / layout.height : 0.75;
  const fitWidth = Math.max(
    1,
    Math.min(available.width - 24, (available.height - 24) * aspect),
  );
  const factor =
    zoom === "fit"
      ? 1
      : zoom === "block" && draft
        ? Math.min(
            8,
            Math.max(
              1,
              (available.width - 24) / (fitWidth * (draft.bounds.width + 0.04)),
            ),
          )
        : zoom === "block"
          ? 1
          : Number(zoom) / 100;
  const width = fitWidth * factor,
    height = width / aspect;
  useEffect(() => {
    if (clicking) setZoom("block");
    else setZoom("fit");
  }, [clicking]);
  useEffect(() => {
    const node = stage.current;
    if (!node) return;
    const frame = requestAnimationFrame(() =>
      node.scrollTo({
        left:
          zoom === "block" && draft
            ? Math.max(0, width * (draft.bounds.x - 0.02))
            : 0,
        top:
          zoom === "block" && draft
            ? Math.max(0, height * (draft.bounds.y - 0.025))
            : 0,
      }),
    );
    return () => cancelAnimationFrame(frame);
  }, [zoom, width, height, draft?.bounds]);
  const circles = useMemo(() => {
    try {
      return [
        ...(guides
          ? layout.blocks.flatMap((b) =>
              blockCircles(b, layout.width, layout.height),
            )
          : []),
        ...(clicking && draft && draft.points.length === draft.options * 2
          ? blockCircles(draft, layout.width, layout.height)
          : []),
      ];
    } catch {
      return [];
    }
  }, [guides, layout.blocks, layout.width, layout.height, draft, clicking]);
  function draw(r: Rect | null) {
    const node = preview.current;
    if (!node) return;
    node.style.visibility = r ? "visible" : "hidden";
    if (r)
      for (const key of ["x", "y", "width", "height"] as const)
        node.setAttribute(key, String(r[key]));
  }
  function cancel() {
    if (animation.current !== null) cancelAnimationFrame(animation.current);
    animation.current = null;
    const g = gesture.current;
    gesture.current = null;
    keyboardStart.current = null;
    draw(null);
    if (g && surface.current?.hasPointerCapture(g.pointer))
      surface.current.releasePointerCapture(g.pointer);
  }
  function position(e: PointerEvent<SVGSVGElement>, box: DOMRect): Point {
    return {
      x: Math.max(0, Math.min(1, (e.clientX - box.left) / box.width)),
      y: Math.max(0, Math.min(1, (e.clientY - box.top) / box.height)),
    };
  }
  function rectangle(a: Point, b: Point): Rect {
    return {
      x: Math.min(a.x, b.x),
      y: Math.min(a.y, b.y),
      width: Math.abs(a.x - b.x),
      height: Math.abs(a.y - b.y),
    };
  }
  function down(e: PointerEvent<SVGSVGElement>) {
    if (e.button !== 0 || !layout.width) return;
    e.preventDefault();
    e.currentTarget.focus({ preventScroll: true });
    const box = e.currentTarget.getBoundingClientRect(),
      p = position(e, box);
    if (clicking) {
      onPoint(p);
      return;
    }
    const rect = rectangle(p, p);
    gesture.current = { start: p, rect, box, pointer: e.pointerId };
    e.currentTarget.setPointerCapture(e.pointerId);
    draw(rect);
  }
  function move(e: PointerEvent<SVGSVGElement>) {
    const g = gesture.current;
    if (!g) return;
    g.rect = rectangle(g.start, position(e, g.box));
    if (animation.current === null)
      animation.current = requestAnimationFrame(() => {
        animation.current = null;
        if (gesture.current) draw(gesture.current.rect);
      });
  }
  function up(e: PointerEvent<SVGSVGElement>) {
    const g = gesture.current;
    if (!g) return;
    const rect = rectangle(g.start, position(e, g.box));
    cancel();
    onBox(rect);
  }
  function keyboard(e: KeyboardEvent<SVGSVGElement>) {
    if (e.key === "Escape") {
      e.preventDefault();
      cancel();
      return;
    }
    if (!layout.width) return;
    const p = cursor ?? { x: 0.5, y: 0.5 };
    if (["ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown"].includes(e.key)) {
      e.preventDefault();
      const n = e.shiftKey ? 50 : e.altKey ? 1 : 5;
      const next = {
        x: Math.max(
          0,
          Math.min(
            1,
            p.x +
              (e.key === "ArrowLeft" ? -n : e.key === "ArrowRight" ? n : 0) /
                layout.width,
          ),
        ),
        y: Math.max(
          0,
          Math.min(
            1,
            p.y +
              (e.key === "ArrowUp" ? -n : e.key === "ArrowDown" ? n : 0) /
                layout.height,
          ),
        ),
      };
      setCursor(next);
      if (keyboardStart.current) draw(rectangle(keyboardStart.current, next));
    } else if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      setCursor(p);
      if (clicking) onPoint(p);
      else if (keyboardStart.current) {
        const rect = rectangle(keyboardStart.current, p);
        cancel();
        onBox(rect);
      } else keyboardStart.current = p;
    }
  }
  return (
    <div className="omr-sheet-column">
      <div className="omr-sheet-toolbar">
        <Field name="Layout zoom">
          <select value={zoom} onChange={(e) => setZoom(e.target.value)}>
            <option value="fit">Fit whole sheet</option>
            <option value="150">150%</option>
            <option value="200">200%</option>
            <option value="300">300%</option>
            <option value="400">400%</option>
            {clicking && <option value="block">Selected answer block</option>}
          </select>
        </Field>
        <span className="muted">
          {clicking
            ? "Click the first two rows"
            : "Draw a frame, then label it"}
        </span>
      </div>
      <div ref={stage} className="omr-sheet-stage">
        <div className="omr-sheet-page" style={{ width, height }}>
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            ref={imageRef}
            src={url || undefined}
            alt="Sheet for guided OMR layout setup"
            draggable={false}
            onLoad={(e) =>
              onLoad(
                e.currentTarget.naturalWidth,
                e.currentTarget.naturalHeight,
              )
            }
            onError={onError}
          />
          <svg
            ref={surface}
            viewBox="0 0 1 1"
            preserveAspectRatio="none"
            tabIndex={0}
            role="group"
            aria-label="Template calibration canvas"
            aria-describedby={help}
            onPointerDown={down}
            onPointerMove={move}
            onPointerUp={up}
            onPointerCancel={cancel}
            onKeyDown={keyboard}
            onFocus={() => setCursor({ x: 0.5, y: 0.5 })}
            onBlur={() => {
              setCursor(null);
              cancel();
            }}
          >
            {Object.entries(layout.identity).map(([kind, r]) => (
              <g key={kind}>
                <rect {...r} className="omr-identity-frame">
                  <title>{fieldName(kind)}</title>
                </rect>
                <text
                  x={r.x}
                  y={Math.max(0.012, r.y - 0.005)}
                  fontSize=".018"
                  fill="#3f6f83"
                >
                  {fieldName(kind)}
                </text>
              </g>
            ))}
            {layout.blocks.map((b, i) => (
              <g key={i}>
                <rect {...b.bounds} className="omr-block-frame">
                  <title>
                    Answers {b.first}–{b.first + b.count - 1}
                  </title>
                </rect>
                <text
                  x={b.bounds.x}
                  y={Math.max(0.012, b.bounds.y - 0.005)}
                  fontSize=".018"
                  fill="#3f6f83"
                >
                  {b.first}–{b.first + b.count - 1}
                </text>
              </g>
            ))}
            {pending && <rect {...pending} className="omr-pending-frame" />}
            {circles.flatMap((row) =>
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
            {clicking &&
              draft?.points.map((p, i) => (
                <ellipse
                  key={i}
                  cx={p.x}
                  cy={p.y}
                  rx={0.006}
                  ry={0.006 * aspect}
                  className="omr-seed-circle"
                />
              ))}
            <rect
              ref={preview}
              x="0"
              y="0"
              width="0"
              height="0"
              style={{ visibility: "hidden" }}
              className="omr-drawing-frame"
            />
            {cursor && (
              <g stroke="#3f6f83" strokeWidth=".002" pointerEvents="none">
                <path
                  d={`M${cursor.x - 0.012} ${cursor.y}h.024 M${cursor.x} ${cursor.y - 0.009}v.018`}
                />
              </g>
            )}
          </svg>
        </div>
      </div>
      <details className="omr-keyboard-help">
        <summary>Keyboard controls</summary>
        <p id={help}>
          Focus the sheet. Arrows move; Enter places a corner or circle. Shift
          moves faster, Alt moves one pixel. Escape cancels a box.
        </p>
      </details>
    </div>
  );
}

export function OMRTemplateBuilder({
  file,
  onProfile,
  onSave,
}: {
  file: File | null;
  onProfile: (profile: OMRProfile | null) => void;
  onSave?: (profile: OMRProfile) => Promise<void>;
}) {
  const [source, setSource] = useState<File | null>(null),
    [url, setUrl] = useState("");
  const [layout, setLayout] = useState<DesignerLayout>(empty),
    [profile, setProfile] = useState<OMRProfile | null>(null);
  const [pending, setPending] = useState<Rect | null>(null),
    [kind, setKind] = useState("");
  const [draft, setDraft] = useState<AnswerBlock | null>(null),
    [editing, setEditing] = useState<number | null>(null);
  const [clicking, setClicking] = useState(false),
    [guides, setGuides] = useState(false),
    [expanded, setExpanded] = useState(false);
  const [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const applied = useRef(false),
    image = useRef<HTMLImageElement | null>(null),
    dialog = useRef<HTMLDialogElement>(null),
    expandButton = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    if (file && !applied.current) {
      setSource(file);
      setLayout(empty());
      resetFrame();
      setError("");
    }
  }, [file]);
  useEffect(() => {
    if (!source) return;
    const next = URL.createObjectURL(source);
    setUrl(next);
    return () => URL.revokeObjectURL(next);
  }, [source]);
  useEffect(() => {
    if (expanded) dialog.current?.showModal();
    else if (dialog.current?.open) {
      dialog.current.close();
      expandButton.current?.focus({ preventScroll: true });
    }
  }, [expanded]);
  const rows = useMemo(
    () =>
      layout.blocks.flatMap((b) =>
        blockCircles(b, layout.width, layout.height),
      ),
    [layout.blocks, layout.width, layout.height],
  );
  const total = layout.blocks.reduce((sum, b) => sum + b.count, 0);
  const draftError = useMemo(() => {
    if (!draft || draft.points.length !== draft.options * 2) return "";
    try {
      nativeFiles({
        ...layout,
        blocks:
          editing === null
            ? [...layout.blocks, draft]
            : layout.blocks.map((b, i) => (i === editing ? draft : b)),
      });
      return "";
    } catch (e) {
      return e instanceof Error ? e.message : "Check this answer block.";
    }
  }, [draft, layout, editing]);
  function resetFrame() {
    setPending(null);
    setKind("");
    setDraft(null);
    setEditing(null);
    setClicking(false);
    setError("");
  }
  function invalidate() {
    applied.current = false;
    setProfile(null);
    onProfile(null);
    setError("");
  }
  function box(r: Rect) {
    if (busy) return;
    if (pending) {
      setError("Save or discard the selected frame before drawing another.");
      return;
    }
    if (r.width * layout.width < 5 || r.height * layout.height < 5) {
      setError("Draw a larger frame around the field.");
      return;
    }
    setPending(r);
    setKind("");
    setError("");
  }
  function choose(value: string) {
    setKind(value);
    setClicking(false);
    setError("");
    if (value === "answers" && pending)
      setDraft({
        bounds: pending,
        points: [],
        first: Math.max(1, ...layout.blocks.map((b) => b.first + b.count)),
        count: 1,
        options: 4,
        size: 0.38,
        labels: "letters",
      });
    else setDraft(null);
  }
  function addPoint(p: Point) {
    if (busy) return;
    if (!draft || draft.points.length >= draft.options * 2) return;
    const next = { ...draft, points: [...draft.points, p] };
    if (next.points.length === next.options * 2)
      try {
        next.count = autoCount(next, layout.width, layout.height);
      } catch {
        /* Keep the clicked points available for undo. */
      }
    setDraft(next);
    setError("");
  }
  function saveFrame() {
    if (!pending) return;
    if (kind === "student_name" || kind === "roll_number") {
      setLayout({
        ...layout,
        identity: { ...layout.identity, [kind]: pending },
      });
      resetFrame();
      return;
    }
    if (draft)
      try {
        const next = {
          ...layout,
          blocks:
            editing === null
              ? [...layout.blocks, draft]
              : layout.blocks.map((b, i) => (i === editing ? draft : b)),
        };
        nativeFiles(next);
        setLayout(next);
        resetFrame();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Check the answer block.");
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
      let blob = await toBlob(
        source?.type === "image/jpeg" ? "image/jpeg" : "image/png",
      );
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
      if (onSave) await onSave(next);
      applied.current = true;
      setProfile(next);
      onProfile(next);
      setExpanded(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Could not create the layout.");
    } finally {
      setBusy(false);
    }
  }
  const editor = (
    <div className="omr-frame-workspace">
      <Sheet
        key={url}
        url={url}
        layout={layout}
        draft={draft}
        pending={pending}
        clicking={clicking}
        guides={guides}
        onBox={box}
        onPoint={addPoint}
        imageRef={image}
        onError={() => setError("Choose a readable PNG or JPEG sheet.")}
        onLoad={(width, height) => {
          if (
            width < 100 ||
            height < 100 ||
            width > 5000 ||
            height > 5000 ||
            width * height > 8_000_000
          ) {
            setError(
              "Use a sheet between 100 pixels and 8 megapixels, with each side no larger than 5000 pixels.",
            );
            return;
          }
          setLayout((old) =>
            old.width === width && old.height === height
              ? old
              : { ...old, width, height },
          );
        }}
      />
      <aside className="omr-frame-inspector" aria-label="OMR frame settings">
        {pending ? (
          <>
            <h4>Label this frame</h4>
            <Field name="Field label">
              <select
                value={kind}
                disabled={editing !== null}
                onChange={(e) => choose(e.target.value)}
              >
                <option value="">Choose what is inside</option>
                <option value="roll_number">Registration number</option>
                <option value="student_name">Student name</option>
                <option value="answers">Answers</option>
              </select>
            </Field>
            {draft && (
              <>
                <p className="muted">
                  One frame per answer column. Set its options, then mark the
                  first two rows.
                </p>
                <div className="omr-frame-pair">
                  <Field name="Options per question">
                    <select
                      value={draft.options}
                      disabled={clicking}
                      onChange={(e) =>
                        setDraft({
                          ...draft,
                          options: Number(e.target.value),
                          points: [],
                        })
                      }
                    >
                      {[2, 3, 4, 5, 6, 7, 8].map((n) => (
                        <option key={n} value={n}>
                          {n}
                        </option>
                      ))}
                    </select>
                  </Field>
                  <Field name="Option labels">
                    <select
                      value={draft.labels ?? "letters"}
                      onChange={(e) =>
                        setDraft({
                          ...draft,
                          labels: e.target.value as "letters" | "numbers",
                        })
                      }
                    >
                      <option value="letters">A, B, C, D</option>
                      <option value="numbers">1, 2, 3, 4</option>
                    </select>
                  </Field>
                </div>
                {!clicking ? (
                  <button
                    type="button"
                    className="primary"
                    onClick={() => setClicking(true)}
                  >
                    Mark first two rows
                  </button>
                ) : (
                  <>
                    <p role="status" className="omr-frame-instruction">
                      {draft.points.length < draft.options * 2
                        ? `Row ${Math.floor(draft.points.length / draft.options) + 1}: click ${optionValue(draft, draft.points.length % draft.options)}. ${draft.points.length} of ${draft.options * 2} selected.`
                        : "Circles filled in. Check the last row and question count."}
                    </p>
                    <button
                      type="button"
                      disabled={!draft.points.length}
                      onClick={() =>
                        setDraft({
                          ...draft,
                          points: draft.points.slice(0, -1),
                        })
                      }
                    >
                      Undo last circle
                    </button>
                  </>
                )}
                {draft.points.length === draft.options * 2 && (
                  <>
                    <button
                      type="button"
                      onClick={() => setDraft({ ...draft, points: [] })}
                    >
                      Re-mark circles
                    </button>
                    <div className="omr-frame-pair">
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
                    </div>
                    <details>
                      <summary>Adjust bubble size</summary>
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
                  </>
                )}
              </>
            )}
            <div className="actions">
              <button
                type="button"
                className="primary"
                disabled={
                  !kind ||
                  (kind === "answers" &&
                    (!draft ||
                      draft.points.length !== draft.options * 2 ||
                      !!draftError))
                }
                onClick={saveFrame}
              >
                Save frame
              </button>
              <button type="button" onClick={resetFrame}>
                Discard frame
              </button>
            </div>
          </>
        ) : (
          <>
            <h4>Your frames</h4>
            <p className="muted">
              Draw around a field on the sheet, then label it here. Add each
              answer column separately.
            </p>
            {!Object.keys(layout.identity).length && !layout.blocks.length && (
              <p className="omr-empty-frames">No frames yet.</p>
            )}
            {Object.entries(layout.identity).map(([key]) => (
              <div className="omr-frame-item" key={key}>
                <strong>{fieldName(key)}</strong>
                <div className="actions">
                  <button
                    type="button"
                    aria-label={`Remove ${fieldName(key).toLowerCase()}`}
                    title={`Remove ${fieldName(key).toLowerCase()}`}
                    onClick={() => {
                      const identity = { ...layout.identity };
                      delete identity[key as IdentityKind];
                      setLayout({ ...layout, identity });
                    }}
                  >
                    ×
                  </button>
                </div>
              </div>
            ))}
            {layout.blocks.map((b, i) => (
              <div className="omr-frame-item" key={i}>
                <strong>
                  Answers {b.first}–{b.first + b.count - 1}
                </strong>
                <span className="muted">
                  {b.count} questions · {b.options} options
                </span>
                <div className="actions">
                  <button
                    type="button"
                    aria-label={`Edit block ${i + 1}`}
                    onClick={() => {
                      setEditing(i);
                      setPending(b.bounds);
                      setKind("answers");
                      setDraft(b);
                      setClicking(true);
                    }}
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    aria-label={`Remove block ${i + 1}`}
                    title={`Remove block ${i + 1}`}
                    onClick={() =>
                      setLayout({
                        ...layout,
                        blocks: layout.blocks.filter((_, j) => j !== i),
                      })
                    }
                  >
                    ×
                  </button>
                </div>
              </div>
            ))}
            {!!layout.blocks.length && (
              <label className="check">
                <input
                  type="checkbox"
                  checked={guides}
                  onChange={(e) => setGuides(e.target.checked)}
                />
                Show bubble guides
              </label>
            )}
            <Field name="OMR layout name">
              <input
                value={layout.name}
                maxLength={80}
                onChange={(e) => setLayout({ ...layout, name: e.target.value })}
              />
            </Field>
            <button
              type="button"
              className="primary"
              disabled={busy || !layout.blocks.length || !layout.name.trim()}
              onClick={() => void create()}
            >
              {busy
                ? "Saving template…"
                : onSave
                  ? "Save template"
                  : "Use this layout"}
            </button>
          </>
        )}
        {(error || draftError) && (
          <p role="alert" className="error">
            {error || draftError}
          </p>
        )}
      </aside>
    </div>
  );
  return (
    <section
      className="omr-designer"
      aria-label="Guided OMR layout setup"
      aria-busy={busy}
    >
      <fieldset disabled={busy} className="omrchecker-inputs">
        <div className="section-heading">
          <div>
            <h3>{profile ? "Your layout is ready" : "Visual OMR template"}</h3>
            <p className="muted">
              Draw a frame. Label it. Repeat for each field.
            </p>
          </div>
          <div className="actions">
            {source && !profile && (
              <button
                ref={expandButton}
                type="button"
                onClick={() => setExpanded(true)}
              >
                Expand editor
              </button>
            )}
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
                  setBusy(true);
                  try {
                    if (f.size > 12_000_000)
                      throw Error("Saved layout is too large.");
                    const next = loadProfile(await f.text());
                    if (onSave) await onSave(next);
                    setLayout(next.layout);
                    setSource(referenceFile(next));
                    resetFrame();
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
                  } finally {
                    setBusy(false);
                  }
                  e.target.value = "";
                }}
              />
            </label>
          </div>
        </div>
        {profile ? (
          <div className="omr-layout-ready">
            <strong>
              {layout.name} · {total} questions · {layout.blocks.length} answer
              blocks
            </strong>
            <p>
              Template, reference and tuning are ready. No marker upload is
              needed.
            </p>
            <div className="actions">
              <button
                ref={expandButton}
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
                  resetFrame();
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
                  resetFrame();
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
        ) : !source ? (
          <p>Upload a sheet image above, or load a saved layout.</p>
        ) : !expanded ? (
          editor
        ) : null}
        <dialog
          ref={dialog}
          className="omr-editor-dialog"
          onCancel={() => setExpanded(false)}
          onClose={() => setExpanded(false)}
        >
          {expanded && !profile && (
            <>
              <div className="omr-dialog-heading">
                <h3>Visual OMR template</h3>
                <button type="button" onClick={() => setExpanded(false)}>
                  Close expanded editor
                </button>
              </div>
              {editor}
            </>
          )}
        </dialog>
        {profile && error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
      </fieldset>
    </section>
  );
}
