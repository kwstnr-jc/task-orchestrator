import { Droppable, Draggable } from "@hello-pangea/dnd";
import type { Project, Task, TaskState } from "../types/task";
import { STATE_LABELS } from "../types/task";
import TaskCard from "./TaskCard";

interface ColumnProps {
  state: TaskState;
  tasks: Task[];
  projects: Project[];
  onEdit: (task: Task) => void;
}

export default function Column({ state, tasks, projects, onEdit }: ColumnProps) {
  return (
    <Droppable droppableId={state}>
      {(provided, snapshot) => (
        <div
          ref={provided.innerRef}
          {...provided.droppableProps}
          className={`flex-shrink-0 w-full md:w-80 bg-gray-900 rounded-lg border transition-colors ${
            snapshot.isDraggingOver
              ? "border-blue-500 bg-gray-800"
              : "border-gray-800"
          }`}
        >
          <div className="sticky top-0 bg-gray-900 border-b border-gray-800 px-4 py-3 z-10 hidden md:block">
            <h2 className="font-semibold text-gray-100">{STATE_LABELS[state]}</h2>
            <p className="text-xs text-gray-500">{tasks.length} tasks</p>
          </div>
          <div className="p-3 space-y-2 md:max-h-[calc(100vh-10rem)] overflow-y-auto">
            {tasks.map((task, index) => (
              <Draggable key={task.id} draggableId={task.id} index={index}>
                {(provided, snapshot) => (
                  <div
                    ref={provided.innerRef}
                    {...provided.draggableProps}
                    {...provided.dragHandleProps}
                  >
                    <TaskCard
                      task={task}
                      project={projects.find((p) => p.id === task.project_id)}
                      isDragging={snapshot.isDragging}
                      onEdit={onEdit}
                    />
                  </div>
                )}
              </Draggable>
            ))}
            {provided.placeholder}
          </div>
        </div>
      )}
    </Droppable>
  );
}
