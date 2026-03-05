import type { Task } from "../types/task";
import { deleteTask, enqueueTask } from "../lib/api";

interface TaskCardProps {
  task: Task;
  isDragging: boolean;
  onRefresh: () => void;
}

export default function TaskCard({ task, isDragging, onRefresh }: TaskCardProps) {
  const handleDelete = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm("Delete this task?")) return;
    try {
      await deleteTask(task.id);
      onRefresh();
    } catch (err) {
      console.error("Failed to delete task:", err);
    }
  };

  const handleEnqueue = async (e: React.MouseEvent) => {
    e.stopPropagation();
    try {
      await enqueueTask(task.id);
      alert("Task enqueued for processing");
    } catch (err) {
      console.error("Failed to enqueue task:", err);
    }
  };

  return (
    <div
      className={`bg-gray-800 border border-gray-700 rounded-lg p-3 cursor-grab transition-shadow ${
        isDragging ? "shadow-lg shadow-blue-500/20 border-blue-500/50" : "hover:border-gray-600"
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <h3 className="text-sm font-medium text-gray-100 leading-tight">
          {task.title}
        </h3>
        <div className="flex gap-1 flex-shrink-0">
          {task.state === "approved" && (
            <button
              onClick={handleEnqueue}
              title="Enqueue for processing"
              className="text-gray-500 hover:text-blue-400 transition-colors"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              </svg>
            </button>
          )}
          <button
            onClick={handleDelete}
            title="Delete task"
            className="text-gray-500 hover:text-red-400 transition-colors"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>
      {task.description && (
        <p className="text-xs text-gray-400 mt-1 line-clamp-2">{task.description}</p>
      )}
      <div className="flex items-center gap-2 mt-2">
        <span
          className={`text-xs px-1.5 py-0.5 rounded ${
            task.task_type === "dev"
              ? "bg-purple-900/50 text-purple-300"
              : "bg-teal-900/50 text-teal-300"
          }`}
        >
          {task.task_type}
        </span>
        {task.priority > 0 && (
          <span className="text-xs text-yellow-400">P{task.priority}</span>
        )}
      </div>
    </div>
  );
}
