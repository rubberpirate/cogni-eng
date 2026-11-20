import type { A } from "./academic";
export type Point = A<"AcademicPoint">;
export type Rect = A<"AcademicGridRect">;
export type IdentityKind = "student_name" | "roll_number";
export type AnswerBlock = {
  first: number;
  count: number;
  options: number;
  bounds: Rect;
  points: Point[];
  size: number;
  labels?: "letters" | "numbers";
};
export type DesignerLayout = {
  width: number;
  height: number;
  name: string;
  identity: Partial<Record<IdentityKind, Rect>>;
  blocks: AnswerBlock[];
};
export type OMRProfile = {
  version: 1;
  layout: DesignerLayout;
  template: Record<string, unknown>;
  config: Record<string, unknown>;
  reference: { name: string; type: string; data: string };
};
export function optionValue(block: AnswerBlock, index: number) {
  return block.labels === "numbers"
    ? String(index + 1)
    : String.fromCharCode(65 + index);
}
export function inside(p: Point, r: Rect) {
  return (
    p.x >= r.x - 1e-8 &&
    p.y >= r.y - 1e-8 &&
    p.x <= r.x + r.width + 1e-8 &&
    p.y <= r.y + r.height + 1e-8
  );
}
export function validRect(r: Rect) {
  return (
    r &&
    [r.x, r.y, r.width, r.height].every(Number.isFinite) &&
    r.x >= 0 &&
    r.y >= 0 &&
    r.width > 0 &&
    r.height > 0 &&
    r.x + r.width <= 1 + 1e-8 &&
    r.y + r.height <= 1 + 1e-8
  );
}

export function pitch(block: AnswerBlock, width: number, height: number) {
  if (
    !Array.isArray(block.points) ||
    block.points.length !== block.options * 2 ||
    block.points.some(
      (p) =>
        !p ||
        !Number.isFinite(p.x) ||
        !Number.isFinite(p.y) ||
        p.x < 0 ||
        p.y < 0 ||
        p.x > 1 ||
        p.y > 1,
    )
  )
    throw Error(
      `Click every bubble in the first two rows (${block.options * 2} clicks).`,
    );
  const rows = [
    block.points.slice(0, block.options),
    block.points.slice(block.options),
  ];
  const mean = (items: number[]) =>
    items.reduce((s, n) => s + n, 0) / items.length;
  const dx = mean(
    rows.map(
      (row) => ((row.at(-1)!.x - row[0].x) * width) / (block.options - 1),
    ),
  );
  const rowX = mean(rows[1].map((p) => p.x)) - mean(rows[0].map((p) => p.x));
  const rowY =
    (mean(rows[1].map((p) => p.y)) - mean(rows[0].map((p) => p.y))) * height;
  const x = mean(rows[0].map((p, i) => p.x * width - i * dx));
  const y = mean(rows[0].map((p) => p.y * height));
  if (dx < 5 || rowY < 5)
    throw Error(
      "Click the first row left to right, then the row directly below it.",
    );
  const box = Math.round(Math.min(dx, rowY) * block.size);
  if (box < 4 || box > 512)
    throw Error(
      "The bubble size is outside the supported range. Adjust the bubble size or use a clearer sheet.",
    );
  for (let row = 0; row < 2; row++)
    for (let i = 0; i < block.options; i++) {
      const p = rows[row][i];
      if (
        !inside(p, block.bounds) ||
        Math.abs(p.x * width - (x + dx * i + rowX * width * row)) >
          box * 0.65 ||
        Math.abs(p.y * height - (y + rowY * row)) > box * 0.65
      )
        throw Error(
          "The clicked circles do not form two even rows. Undo and click their centers again.",
        );
    }
  return { x, y, dx, rowX: rowX * width, rowY, box };
}
export function blockCircles(
  block: AnswerBlock,
  width: number,
  height: number,
) {
  if (!Number.isInteger(block.count) || block.count < 1 || block.count > 200)
    throw Error("Use between 1 and 200 questions in each answer box.");
  const p = pitch(block, width, height);
  return Array.from({ length: block.count }, (_, row) => ({
    number: block.first + row,
    points: Array.from({ length: block.options }, (_, col) => ({
      x: (p.x + p.rowX * row + p.dx * col) / width,
      y: (p.y + p.rowY * row) / height,
    })),
    size: p.box,
  }));
}
export function autoCount(block: AnswerBlock, width: number, height: number) {
  const p = pitch(block, width, height);
  return Math.max(
    1,
    Math.floor(
      ((block.bounds.y + block.bounds.height) * height - p.box / 2 - p.y) /
        p.rowY,
    ) + 1,
  );
}
export function nativeFiles(
  layout: DesignerLayout,
  referenceName = "reference.png",
) {
  if (
    typeof layout.name !== "string" ||
    !layout.name.trim() ||
    layout.name.length > 80
  )
    throw Error("Give this layout a short name.");
  if (
    !Number.isInteger(layout.width) ||
    !Number.isInteger(layout.height) ||
    layout.width < 100 ||
    layout.height < 100 ||
    layout.width > 5000 ||
    layout.height > 5000 ||
    layout.width * layout.height > 8_000_000
  )
    throw Error(
      "Use a sheet between 100 pixels and 8 megapixels, with each side no larger than 5000 pixels.",
    );
  if (!layout.blocks.length || layout.blocks.length > 8)
    throw Error("Add one to eight answer blocks.");
  const fields: Record<string, unknown> = {},
    seen = new Set<number>();
  for (const block of layout.blocks) {
    if (
      !Number.isInteger(block.options) ||
      block.options < 2 ||
      block.options > 8 ||
      !Number.isInteger(block.count) ||
      block.count < 1 ||
      block.count > 200 ||
      !Number.isSafeInteger(block.first) ||
      block.first < 1 ||
      !Number.isSafeInteger(block.first + block.count) ||
      !validRect(block.bounds) ||
      !Number.isFinite(block.size) ||
      (block.labels !== undefined &&
        !["letters", "numbers"].includes(block.labels)) ||
      block.size < 0.15 ||
      block.size > 0.8
    )
      throw Error(
        "Check each block’s question numbers, count and bubble size.",
      );
    for (const row of blockCircles(block, layout.width, layout.height)) {
      if (seen.has(row.number))
        throw Error(`Question ${row.number} appears in more than one block.`);
      seen.add(row.number);
      if (seen.size > 500) throw Error("Use at most 500 questions per layout.");
      const first = row.points[0],
        last = row.points.at(-1)!;
      const origin = [
        Math.round(first.x * layout.width - row.size / 2),
        Math.round(first.y * layout.height - row.size / 2),
      ];
      if (
        !inside(
          {
            x: first.x - row.size / 2 / layout.width,
            y: first.y - row.size / 2 / layout.height,
          },
          block.bounds,
        ) ||
        !inside(
          {
            x: last.x + row.size / 2 / layout.width,
            y: last.y + row.size / 2 / layout.height,
          },
          block.bounds,
        ) ||
        origin[0] < 0 ||
        origin[1] < 0
      )
        throw Error(
          `Question ${row.number} leaves its answer box. Check the row count or redraw the box.`,
        );
      const p = pitch(block, layout.width, layout.height);
      fields[`answer_${row.number}`] = {
        origin,
        fieldLabels: [`q${row.number}`],
        bubbleValues: Array.from({ length: block.options }, (_, i) =>
          optionValue(block, i),
        ),
        direction: "horizontal",
        bubbleDimensions: [row.size, row.size],
        bubblesGap: p.dx,
        labelsGap: 0,
      };
    }
  }
  for (const [kind, r] of Object.entries(layout.identity))
    if (!["student_name", "roll_number"].includes(kind) || !validRect(r))
      throw Error("Identity boxes must stay inside the sheet.");
  const scale = Math.min(1, 2000 / Math.max(layout.width, layout.height));
  return {
    template: {
      pageDimensions: [layout.width, layout.height],
      bubbleDimensions: [20, 20],
      fieldBlocks: fields,
      preProcessors: [
        {
          name: "FeatureBasedAlignment",
          options: {
            reference: referenceName,
            maxFeatures: 2500,
            goodMatchPercent: 0.3,
          },
        },
      ],
    },
    config: {
      dimensions: {
        processing_width: Math.max(100, Math.round(layout.width * scale)),
        processing_height: Math.max(100, Math.round(layout.height * scale)),
      },
      alignment_params: { auto_align: false },
      outputs: {
        show_image_level: 0,
        save_image_level: 0,
        save_detections: false,
      },
    },
  };
}
export function loadProfile(text: string): OMRProfile {
  if (text.length > 12_000_000) throw Error("Saved layout is too large.");
  const p = JSON.parse(text);
  if (
    p.version !== 1 ||
    !p.layout ||
    !Array.isArray(p.layout.blocks) ||
    !p.layout.identity ||
    typeof p.layout.name !== "string" ||
    !p.reference ||
    !["image/png", "image/jpeg"].includes(p.reference.type) ||
    typeof p.reference.data !== "string" ||
    !/^[-A-Za-z0-9+/=]+$/.test(p.reference.data) ||
    !["reference.png", "reference.jpg"].includes(p.reference.name)
  )
    throw Error("Choose a saved .omr.json layout.");
  const generated = nativeFiles(p.layout, p.reference.name);
  return { version: 1, layout: p.layout, ...generated, reference: p.reference };
}
export function referenceFile(profile: OMRProfile) {
  return new File(
    [Uint8Array.from(atob(profile.reference.data), (c) => c.charCodeAt(0))],
    profile.reference.name,
    { type: profile.reference.type },
  );
}
