import { useState } from "react";
import type { Project, Task, TaskState } from "../types/task";
import { STATE_LABELS } from "../types/task";
import { updateTask, deleteTask } from "../lib/api";
import Dropdown from "./Dropdown";

interface EditTaskFormProps {
  task: Task;
  projects: Project[];
  states: TaskState[];
  onClose: () => void;
  onUpdated: () => void;
}

export default function EditTaskForm({
  task,
  projects,
  states,
  onClose,
  onUpdated,
}: EditTaskFormProps) {
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description || "");
  const [priority, setPriority] = useState(task.priority);
  const [projectId, setProjectId] = useState(task.project_id || "");
  const [loading, setLoading] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const currentIdx = states.indexOf(task.state);
  const prevState = currentIdx > 0 ? states[currentIdx - 1] : null;
  const nextState = currentIdx < states.length - 1 ? states[currentIdx + 1] : null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    setLoading(true);
    try {
      await updateTask(task.id, {
        title: title.trim(),
        description: description.trim() || undefined,
        priority: Number(priority),
        project_id: projectId || null,
      });
      onUpdated();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleStateChange = async (newState: TaskState) => {
    setLoading(true);
    try {
      await updateTask(task.id, { state: newState });
      onUpdated();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    setLoading(true);
    try {
      await deleteTask(task.id);
      onUpdated();
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  if (showDeleteConfirm) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-50 flex items-end lg:items-center justify-center z-50 lg:p-4">
        <div className="bg-gray-900 border border-gray-800 rounded-t-2xl lg:rounded-lg p-6 w-full max-w-96 shadow-xl">
          <h2 className="text-lg font-bold mb-2">Delete Task</h2>
          <p className="text-gray-400 text-sm mb-6">
            Are you sure you want to delete{" "}
            <span className="text-white font-medium">{task.title}</span>?
          </p>
          <div className="flex gap-2 justify-end">
            <button
              onClick={() => setShowDeleteConfirm(false)}
              className="px-4 py-2 rounded-lg text-sm text-gray-400 hover:text-gray-300 transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              onClick={handleDelete}
              className="px-4 py-2 rounded-lg text-sm bg-red-600 hover:bg-red-500 text-white transition-colors disabled:opacity-50"
              disabled={loading}
            >
              {loading ? "Deleting..." : "Delete"}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-end lg:items-center justify-center z-50 lg:p-4">
      <div className="bg-gray-900 border border-gray-800 rounded-t-2xl lg:rounded-lg p-6 w-full max-w-96 shadow-xl max-h-[90vh] overflow-y-auto">
        <h2 className="text-lg font-bold mb-4">Edit Task</h2>

        {/* State indicator + move buttons */}
        <div className="flex items-center gap-2 mb-4">
          <button
            onClick={() => prevState && handleStateChange(prevState)}
            disabled={!prevState || loading}
            className="px-2 py-1 rounded text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-gray-400"
            title={prevState ? `Move to ${STATE_LABELS[prevState]}` : undefined}
          >
            &larr;
          </button>
          <span className="flex-1 text-center text-sm font-medium text-gray-300 bg-gray-800 rounded-lg py-1.5">
            {STATE_LABELS[task.state]}
          </span>
          <button
            onClick={() => nextState && handleStateChange(nextState)}
            disabled={!nextState || loading}
            className="px-2 py-1 rounded text-sm text-gray-400 hover:text-white hover:bg-gray-800 transition-colors disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-gray-400"
            title={nextState ? `Move to ${STATE_LABELS[nextState]}` : undefined}
          >
            &rarr;
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Title
            </label>
            <input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
              autoFocus
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Description
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500 resize-none"
              rows={3}
              disabled={loading}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Project
            </label>
            <Dropdown
              value={projectId}
              onChange={setProjectId}
              placeholder="No project"
              disabled={loading}
              options={[
                { value: "", label: "No project" },
                ...projects.map((p) => ({ value: p.id, label: p.name })),
              ]}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-300 mb-1">
              Priority
            </label>
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(Number(e.target.value))}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
              disabled={loading}
            />
          </div>

          <div className="flex justify-between pt-4">
            <button
              type="button"
              onClick={() => setShowDeleteConfirm(true)}
              className="px-4 py-2 rounded-lg text-sm text-red-500 hover:text-red-400 transition-colors disabled:opacity-50"
              disabled={loading}
            >
              Delete
            </button>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 rounded-lg text-sm text-gray-400 hover:text-gray-300 transition-colors disabled:opacity-50"
                disabled={loading}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-4 py-2 rounded-lg text-sm bg-blue-600 hover:bg-blue-500 text-white transition-colors disabled:opacity-50"
                disabled={loading}
              >
                {loading ? "Saving..." : "Save"}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
