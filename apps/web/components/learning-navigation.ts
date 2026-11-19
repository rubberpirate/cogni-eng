export const learningViews = {
  home: { title: "Home", hint: "Choose what you want to do next." },
  author: {
    title: "Enter questions manually",
    hint: "Write a question, choose the correct answer and set the marks.",
  },
  generate: {
    title: "Create questions with AI",
    hint: "Describe the topic, generate drafts, then check the answers.",
  },
  curriculum: {
    title: "Question bank",
    hint: "Find, write and approve questions before using them in a test.",
  },
  import: {
    title: "Upload questions",
    hint: "Upload a question paper or spreadsheet, check the preview, then save questions.",
  },
  papers: {
    title: "Assessments",
    hint: "Build a test, publish it and choose who should take it.",
  },
  scan: {
    title: "Scan & check",
    hint: "Upload completed answer sheets, check uncertain marks and save results.",
  },
  results: {
    title: "Results & review",
    hint: "Review submitted work, approve written marks and release results.",
  },
  samples: {
    title: "Try sample files",
    hint: "Use an original practice paper and matching answer sheets to try the complete flow.",
  },
  jobs: {
    title: "Background tasks",
    hint: "Check processing, fix a failed request or retry when a provider is available.",
  },
} as const;
export type LearningView = keyof typeof learningViews;
