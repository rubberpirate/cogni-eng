import { test, expect } from "bun:test";
import {
  autoCount,
  blockCircles,
  nativeFiles,
  loadProfile,
  type AnswerBlock,
  type DesignerLayout,
} from "../clients/omr-template";
const block = (first = 1, x = 180): AnswerBlock => ({
  first,
  count: 10,
  options: 4,
  size: 0.38,
  bounds: {
    x: (x - 25) / 1200,
    y: 725 / 1600,
    width: 270 / 1200,
    height: 690 / 1600,
  },
  points: [750, 820].flatMap((y) =>
    [x, x + 70, x + 140, x + 210].map((x) => ({ x: x / 1200, y: y / 1600 })),
  ),
});
const layout = (): DesignerLayout => ({
  width: 1200,
  height: 1600,
  name: "Test layout",
  identity: { student_name: { x: 0.06, y: 0.1, width: 0.7, height: 0.03 } },
  blocks: [block(), block(11, 740)],
});
test("eight clicks interpolate every question and produce native files without requiring markers", () => {
  const l = layout();
  expect(autoCount(l.blocks[0], l.width, l.height)).toBe(10);
  const rows = blockCircles(l.blocks[0], l.width, l.height);
  expect(rows).toHaveLength(10);
  expect(rows[9].points[3].y).toBeCloseTo(1380 / 1600);
  const files = nativeFiles(l);
  expect(Object.keys(files.template.fieldBlocks)).toHaveLength(20);
  expect(files.template.preProcessors[0].name).toBe("FeatureBasedAlignment");
  expect(files.template.fieldBlocks.answer_20).toMatchObject({
    fieldLabels: ["q20"],
    bubbleValues: ["A", "B", "C", "D"],
  });
});
test("ambiguous clicks, duplicate numbering, out-of-box rows and unsafe imports are rejected", () => {
  expect(() => blockCircles({ ...block(), count: 1e9 }, 1200, 1600)).toThrow(
    "between 1 and 200",
  );
  for (const change of [
    (l: DesignerLayout) => l.blocks[0].points.reverse(),
    (l: DesignerLayout) => (l.blocks[1].first = 1),
    (l: DesignerLayout) => (l.blocks[0].count = 200),
  ]) {
    const l = layout();
    change(l);
    expect(() => nativeFiles(l)).toThrow();
  }
  const l = layout();
  const files = nativeFiles(l);
  const saved = {
    version: 1,
    layout: l,
    ...files,
    reference: { name: "reference.png", type: "image/png", data: "AQID" },
  };
  expect(loadProfile(JSON.stringify(saved)).layout.blocks).toHaveLength(2);
  expect(() =>
    loadProfile(
      JSON.stringify({
        ...saved,
        reference: { ...saved.reference, name: "../private.png" },
      }),
    ),
  ).toThrow();
});
