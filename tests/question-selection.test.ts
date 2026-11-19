import { test, expect } from "bun:test";
import { selectQuestionNumbers } from "../apps/web/components/question-selection";
const available = Array.from({ length: 25 }, (_, i) => i + 1);
test("selects inclusive ranges and comma lists without duplicate writes", () => {
  expect(selectQuestionNumbers("1-3, 2,5,8,14,25", available)).toEqual([
    1, 2, 3, 5, 8, 14, 25,
  ]);
});
test("rejects missing, descending, malformed and excessive ranges", () => {
  for (const input of [
    "",
    "0",
    "3-1",
    "1,",
    "1.5",
    "1-99999999999999999999",
    "1-26",
    "2-4",
  ]) {
    expect(() => selectQuestionNumbers(input, [1, 2, 4, 5])).toThrow();
  }
});
