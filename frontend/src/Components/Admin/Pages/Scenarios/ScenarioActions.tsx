import { useEffect, useRef, useState } from "react";
import styles from "./Styles.module.scss";
import { EllipsisVertical } from "lucide-react";
import cn from "classnames";
import { Trash2, Eye, Clock3 } from "lucide-react";

type Props = {
  onEdit: () => void;
  onPublish: () => void;
  onDelete: () => void;
};

export function ScenarioActions({ onEdit, onPublish, onDelete }: Props) {
  const [open, setOpen] = useState(false);
  const [openUp, setOpenUp] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;

    const handleOutsideClick = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) setOpen(false);
    };

    document.addEventListener("mousedown", handleOutsideClick);
    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
    };
  }, [open]);

  const handleAction = (action: () => void) => {
    setOpen(false);
    action();
  };

  const toggleMenu = () => {
    if (!open && menuRef.current) {
      const rect = menuRef.current.getBoundingClientRect();
      setOpenUp(window.innerHeight - rect.bottom < 190);
    }
    setOpen((prev) => !prev);
  };

  return (
    <div className={styles.actions} ref={menuRef}>
      <button
        type="button"
        className={styles.actions__trigger}
        aria-haspopup="menu"
        aria-expanded={open}
        onClick={toggleMenu}
      >
        <EllipsisVertical size={33} color="#94a3b8" />
      </button>

      {open && (
        <div className={cn(styles.actions__menu, openUp && styles.actions__menu_up)} role="menu">
          <button type="button" role="menuitem" onClick={() => handleAction(onEdit)}>
            <Eye size={20} color="#94A3B8" />
            Предпросмотр
          </button>

          <button type="button" role="menuitem" onClick={() => handleAction(onPublish)}>
            <Clock3 size={19} color="#94A3B8" />
            История версий
          </button>

          <button type="button" role="menuitem" className={styles.actions__delete} onClick={() => handleAction(onDelete)}>
            <Trash2 size={19} color="#E11D48" />
            Удалить
          </button>
        </div>
      )}
    </div>
  );
}