export type TaskType = "dev" | "research";
export type TaskState = "draft" | "refine" | "approved" | "in_progress" | "done";

export interface Task {
  id: string;
  title: string;
  description: string | null;
  task_type: TaskType;
  state: TaskState;
  priority: number;
  metadata: Record<string, unknown>;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface UserInfo {
  username: string;
  display_name: string;
  source: string;
}

export const DEV_STATES: TaskState[] = [
  "draft",
  "refine",
  "approved",
  "in_progress",
  "done",
];

export const RESEARCH_STATES: TaskState[] = [
  "draft",
  "in_progress",
  "done",
];

export const STATE_LABELS: Record<TaskState, string> = {
  draft: "Draft",
  refine: "Refine",
  approved: "Approved",
  in_progress: "In Progress",
  done: "Done",
};
