import { useCallback, useEffect, useState } from "react";
import {
  DragDropContext,
  type DropResult,
} from "@hello-pangea/dnd";
import type { Task, TaskType, UserInfo } from "../types/task";
import { DEV_STATES, RESEARCH_STATES } from "../types/task";
import { getTasks, updateTask } from "../lib/api";
import Column from "../components/Column";
import CreateTaskForm from "../components/CreateTaskForm";

interface BoardPageProps {
  user: UserInfo;
  onLogout: () => void;
}

export default function BoardPage({ user, onLogout }: BoardPageProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [activeType, setActiveType] = useState<TaskType>("dev");
  const [showCreate, setShowCreate] = useState(false);

  const loadTasks = useCallback(() => {
    getTasks({ type: activeType }).then(setTasks).catch(console.error);
  }, [activeType]);

  useEffect(() => {
    loadTasks();
  }, [loadTasks]);

  const states = activeType === "dev" ? DEV_STATES : RESEARCH_STATES;

  const handleDragEnd = async (result: DropResult) => {
    if (!result.destination) return;
    const { draggableId, destination } = result;
    const newState = destination.droppableId;

    const task = tasks.find((t) => t.id === draggableId);
    if (!task || task.state === newState) return;

    // Optimistic update
    setTasks((prev) =>
      prev.map((t) =>
        t.id === draggableId ? { ...t, state: newState as Task["state"] } : t
      )
    );

    try {
      await updateTask(draggableId, { state: newState as Task["state"] });
    } catch {
      loadTasks(); // Revert on error
    }
  };

  const handleLogout = async () => {
    await fetch("/api/auth/logout", { credentials: "include" });
    onLogout();
  };

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <h1 className="text-xl font-bold">Task Orchestrator</h1>
          <div className="flex gap-1 bg-gray-900 rounded-lg p-1">
            <button
              onClick={() => setActiveType("dev")}
              className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                activeType === "dev"
                  ? "bg-gray-700 text-white"
                  : "text-gray-400 hover:text-gray-200"
              }`}
            >
              Dev
            </button>
            <button
              onClick={() => setActiveType("research")}
              className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                activeType === "research"
                  ? "bg-gray-700 text-white"
                  : "text-gray-400 hover:text-gray-200"
              }`}
            >
              Research
            </button>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <button
            onClick={() => setShowCreate(true)}
            className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
          >
            + New Task
          </button>
          <span className="text-sm text-gray-400">{user.username}</span>
          <button
            onClick={handleLogout}
            className="text-sm text-gray-500 hover:text-gray-300 transition-colors"
          >
            Logout
          </button>
        </div>
      </header>

      {/* Board */}
      <div className="flex-1 p-6 overflow-x-auto">
        <DragDropContext onDragEnd={handleDragEnd}>
          <div className="flex gap-4 min-h-[calc(100vh-8rem)]">
            {states.map((state) => (
              <Column
                key={state}
                state={state}
                tasks={tasks.filter((t) => t.state === state)}
                onRefresh={loadTasks}
              />
            ))}
          </div>
        </DragDropContext>
      </div>

      {/* Create Task Modal */}
      {showCreate && (
        <CreateTaskForm
          taskType={activeType}
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false);
            loadTasks();
          }}
        />
      )}
    </div>
  );
}
