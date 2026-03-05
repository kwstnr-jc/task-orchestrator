import { Droppable, Draggable } from "@hello-pangea/dnd";
import type { Task, TaskState } from "../types/task";
import { STATE_LABELS } from "../types/task";
import TaskCard from "./TaskCard";

interface ColumnProps {
  state: TaskState;
  tasks: Task[];
  onRefresh: () => void;
}

const STATE_COLORS: Record<TaskState, string> = {
  draft: "border-gray-600",
  refine: "border-yellow-600",
  approved: "border-green-600",
  in_progress: "border-blue-600",
  done: "border-emerald-600",
};

export default function Column({ state, tasks, onRefresh }: ColumnProps) {
  return (
    <div className="flex-1 min-w-[260px] max-w-[340px]">
      <div
        className={`flex items-center gap-2 mb-3 pb-2 border-b-2 ${STATE_COLORS[state]}`}
      >
        <h2 className="text-sm font-semibold text-gray-300 uppercase tracking-wide">
          {STATE_LABELS[state]}
        </h2>
        <span className="text-xs text-gray-500 bg-gray-800 px-2 py-0.5 rounded-full">
          {tasks.length}
        </span>
      </div>
      <Droppable droppableId={state}>
        {(provided, snapshot) => (
          <div
            ref={provided.innerRef}
            {...provided.droppableProps}
            className={`space-y-2 min-h-[200px] rounded-lg p-2 transition-colors ${
              snapshot.isDraggingOver ? "bg-gray-800/50" : ""
            }`}
          >
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
                      isDragging={snapshot.isDragging}
                      onRefresh={onRefresh}
                    />
                  </div>
                )}
              </Draggable>
            ))}
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </div>
  );
}
