import { describe, it, expect, vi, beforeEach } from "vitest";

// Mock fetch globally before importing api module
const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

// Import after mock is set up
const api = await import("./api");

describe("API client", () => {
  beforeEach(() => {
    mockFetch.mockReset();
  });

  it("getMe calls /api/auth/me", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ username: "dev", display_name: "Dev User" }),
    });

    const user = await api.getMe();
    expect(user.username).toBe("dev");
    expect(mockFetch).toHaveBeenCalledWith("/api/auth/me", expect.objectContaining({
      credentials: "include",
    }));
  });

  it("getTasks calls /api/tasks with query params", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: () => Promise.resolve([]),
    });

    await api.getTasks({ type: "dev", state: "draft" });
    const url = mockFetch.mock.calls[0][0] as string;
    expect(url).toContain("/tasks");
    expect(url).toContain("type=dev");
    expect(url).toContain("state=draft");
  });

  it("createTask sends POST with body", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: () => Promise.resolve({ id: "123", title: "Test" }),
    });

    await api.createTask({ title: "Test", task_type: "dev" });
    const [, opts] = mockFetch.mock.calls[0];
    expect(opts.method).toBe("POST");
    expect(JSON.parse(opts.body)).toEqual({ title: "Test", task_type: "dev" });
  });

  it("throws on 401 response", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
    });

    await expect(api.getMe()).rejects.toThrow("unauthorized");
  });

  it("deleteTask sends DELETE", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
    });

    await api.deleteTask("some-id");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toBe("/api/tasks/some-id");
    expect(opts.method).toBe("DELETE");
  });

  it("deleteProject sends DELETE", async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      status: 204,
    });

    await api.deleteProject("proj-id");
    const [url, opts] = mockFetch.mock.calls[0];
    expect(url).toBe("/api/projects/proj-id");
    expect(opts.method).toBe("DELETE");
  });
});
