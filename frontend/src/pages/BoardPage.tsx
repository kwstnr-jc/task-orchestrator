import { useCallback, useEffect, useState } from "react";
import {
  DragDropContext,
  type DropResult,
} from "@hello-pangea/dnd";
import type { Project, Task, TaskType, UserInfo } from "../types/task";
import { DEV_STATES, RESEARCH_STATES } from "../types/task";
import { getTasks, getProjects, updateTask, createProject } from "../lib/api";
import Column from "../components/Column";
import CreateTaskForm from "../components/CreateTaskForm";

interface BoardPageProps {
  user: UserInfo;
  onLogout: () => void;
}

export default function BoardPage({ user, onLogout }: BoardPageProps) {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeType, setActiveType] = useState<TaskType>("dev");
  const [selectedProject, setSelectedProject] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [showNewProject, setShowNewProject] = useState(false);
  const [newProjectName, setNewProjectName] = useState("");

  const loadProjects = useCallback(() => {
    getProjects().then(setProjects).catch(console.error);
  }, []);

  const loadTasks = useCallback(() => {
    getTasks({ 
      type: activeType,
      project_id: selectedProject || undefined
    }).then(setTasks).catch(console.error);
  }, [activeType, selectedProject]);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

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

    setTasks((prev) =>
      prev.map((t) =>
        t.id === draggableId ? { ...t, state: newState as Task["state"] } : t
      )
    );

    try {
      await updateTask(draggableId, { state: newState as Task["state"] });
    } catch {
      loadTasks();
    }
  };

  const handleCreateProject = async () => {
    if (!newProjectName.trim()) return;
    try {
      await createProject({ name: newProjectName });
      setNewProjectName("");
      setShowNewProject(false);
      loadProjects();
    } catch (err) {
      console.error(err);
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
          
          {/* Project Selector */}
          <select
            value={selectedProject || ""}
            onChange={(e) => setSelectedProject(e.target.value || null)}
            className="bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-sm text-gray-200 hover:bg-gray-800 transition-colors"
          >
            <option value="">All Projects</option>
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>

          <button
            onClick={() => setShowNewProject(true)}
            className="text-gray-500 hover:text-gray-300 text-sm transition-colors"
            title="New project"
          >
            ⊕ Project
          </button>

          {/* Task Type Toggle */}
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
                projects={projects}
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
          projects={projects}
          selectedProject={selectedProject}
          onClose={() => setShowCreate(false)}
          onCreated={() => {
            setShowCreate(false);
            loadTasks();
          }}
        />
      )}

      {/* Create Project Modal */}
      {showNewProject && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-800 rounded-lg p-6 w-96 shadow-xl">
            <h2 className="text-lg font-bold mb-4">New Project</h2>
            <input
              type="text"
              value={newProjectName}
              onChange={(e) => setNewProjectName(e.target.value)}
              placeholder="Project name"
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-white mb-4 focus:outline-none focus:border-blue-500"
              autoFocus
              onKeyDown={(e) => {
                if (e.key === "Enter") handleCreateProject();
                if (e.key === "Escape") setShowNewProject(false);
              }}
            />
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => setShowNewProject(false)}
                className="px-4 py-2 rounded-lg text-sm text-gray-400 hover:text-gray-300 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleCreateProject}
                className="px-4 py-2 rounded-lg text-sm bg-blue-600 hover:bg-blue-500 text-white transition-colors"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
