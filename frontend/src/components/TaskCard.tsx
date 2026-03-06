import type { Project, Task } from "../types/task";
import { deleteTask } from "../lib/api";

interface TaskCardProps {
  task: Task;
  project?: Project;
  isDragging: boolean;
  onRefresh: () => void;
}

export default function TaskCard({
  task,
  project,
  isDragging,
  onRefresh,
}: TaskCardProps) {
  const handleDelete = async () => {
    if (confirm("Delete this task?")) {
      try {
        await deleteTask(task.id);
        onRefresh();
      } catch (err) {
        console.error(err);
      }
    }
  };

  return (
    <div
      className={`p-3 rounded-lg border transition-all ${
        isDragging
          ? "border-blue-500 bg-gray-800 shadow-lg"
          : "border-gray-700 bg-gray-800 hover:border-gray-600"
      }`}
    >
      <div className="flex items-start justify-between gap-2">
        <div className="flex-1">
          <h3 className="text-sm font-medium text-gray-100 break-words">
            {task.title}
          </h3>
          {task.description && (
            <p className="text-xs text-gray-400 mt-1 line-clamp-2">
              {task.description}
            </p>
          )}
          <div className="flex items-center gap-2 mt-2 flex-wrap">
            {project && (
              <span
                className="inline-block px-2 py-1 text-xs rounded text-white"
                style={{ backgroundColor: project.color }}
              >
                {project.name}
              </span>
            )}
            <span className="text-xs text-gray-500 bg-gray-700 px-2 py-1 rounded">
              {task.task_type}
            </span>
          </div>
        </div>
        <button
          onClick={handleDelete}
          className="text-gray-500 hover:text-red-400 transition-colors flex-shrink-0"
          title="Delete task"
        >
          ✕
        </button>
      </div>
    </div>
  );
}
