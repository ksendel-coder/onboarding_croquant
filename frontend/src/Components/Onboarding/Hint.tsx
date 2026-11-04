import { useEffect, useState } from "react";
import type { Hint as HintType } from "../../Types";
import styles from "./Hint.module.scss";

interface Props {
  hint: HintType;
  index: number;
  total: number;
  onNext: () => void;
  onClose: () => void;
  busy: boolean;
}

export function Hint({ hint, index, total, onNext, onClose, busy }: Props) {
  const [position, setPosition] = useState({ top: 120, left: 24 });

  useEffect(() => {
    const update = () => {
      let element: Element | null = null;

      try { element = document.querySelector(hint.selector); } catch { return; }
      if (!element) return;

      const rect = element.getBoundingClientRect();
      const width = Math.min(340, window.innerWidth - 32);
      const height = 200;
      const centerX = rect.left + rect.width / 2 - width / 2;
      const centerY = rect.top + rect.height / 2 - height / 2;
      let left = centerX;
      let top = rect.bottom + 12;

      switch (hint.placement) {
        case "top": top = rect.top - height - 12; break;
        case "bottom": top = rect.bottom + 12; break;
        case "left": left = rect.left - width - 12; top = centerY; break;
        case "right": left = rect.right + 12; top = centerY; break;
        case "left-top": left = rect.left - width - 12; top = rect.top; break;
        case "right-top": left = rect.right + 12; top = rect.top; break;
        case "center": left = centerX; top = centerY; break;
      }
      setPosition({
        left: Math.max(16, Math.min(left, window.innerWidth - width - 16)),
        top: Math.max(16, Math.min(top, window.innerHeight - height - 16)),
      });
    };
    update();

    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);

    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
    };
  }, [hint.selector, hint.placement]);

  return (
    <aside className={styles.hint} style={position} role="dialog" aria-label={hint.title}>
      <button className={styles.close} type="button" aria-label="Закрыть" onClick={onClose} disabled={busy}>×</button>
      <small>Шаг {index + 1} из {total}</small>
      <h3>{hint.title}</h3>
      <p>{hint.content}</p>
      <button type="button" onClick={onNext} disabled={busy}>{index === total - 1 ? "Завершить" : "Далее"}</button>
    </aside>
  );
}