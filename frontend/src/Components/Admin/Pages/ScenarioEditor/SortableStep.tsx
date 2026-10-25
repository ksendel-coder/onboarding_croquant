import type { ReactNode } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
import styles from "./ScenarioEditor.module.scss";

interface SortableStepProps {
  id: string;
  disabled?: boolean;
  children: ReactNode;
}

export function SortableStep({ id, disabled = false, children }: SortableStepProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id,
    disabled,
  });

  return (
    <li
      ref={setNodeRef}
      className={`${styles.stepItem} ${isDragging ? styles.stepDragging : ""}`}
      style={{ transform: CSS.Transform.toString(transform), transition }}
    >
      <button
        type="button"
        className={styles.dragHandle}
        disabled={disabled}
        aria-label="Перетащить шаг"
        title="Перетащить шаг"
        {...attributes}
        {...listeners}
      >
        <GripVertical size={22} />
      </button>
      
      {children}
    </li>
  );
}
