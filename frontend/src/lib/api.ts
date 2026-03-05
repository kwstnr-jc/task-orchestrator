import type { Task, TaskState, TaskType, UserInfo } from "../types/task";

const BASE = "/api";

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    ...options,
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || res.statusText);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export function getMe(): Promise<UserInfo> {
  return request("/auth/me");
}

export function getTasks(params?: {
  type?: TaskType;
  state?: TaskState;
}): Promise<Task[]> {
  const query = new URLSearchParams();
  if (params?.type) query.set("type", params.type);
  if (params?.state) query.set("state", params.state);
  const qs = query.toString();
  return request(`/tasks${qs ? `?${qs}` : ""}`);
}

export function createTask(data: {
  title: string;
  description?: string;
  task_type: TaskType;
  priority?: number;
}): Promise<Task> {
  return request("/tasks", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function updateTask(
  id: string,
  data: { title?: string; description?: string; state?: TaskState; priority?: number }
): Promise<Task> {
  return request(`/tasks/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export function deleteTask(id: string): Promise<void> {
  return request(`/tasks/${id}`, { method: "DELETE" });
}

export function enqueueTask(id: string): Promise<{ status: string; task_id: string }> {
  return request(`/tasks/${id}/enqueue`, { method: "POST" });
}
