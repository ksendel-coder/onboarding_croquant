
import { useEffect, useState, useRef } from "react";
import type { Hint as HintType } from "../../Types";
import styles from "./Hint.module.scss";

interface Props {
  hint: HintType;
  index: number;
  total: number;
  onNext: () => void;
  onClose: () => void;
  busy: boolean;
  preview?: boolean;
}

interface Position {
  top: number;
  left: number;
}

interface Highlight {
  top: number;
  left: number;
  width: number;
  height: number;
}

export function Hint({ hint, index, total, onNext, onClose, busy, preview = false }: Props) {
  const [position, setPosition] = useState<Position | null>(null);
  const [highlight, setHighlight] = useState<Highlight | null>(null);
  const onNextRef = useRef(onNext);

  useEffect(() => { onNextRef.current = onNext; }, [onNext]);

  useEffect(() => {
    const update = () => {
      let element: Element | null = null;

      try {
        if (hint.selector) {
          const root = preview ? document.querySelector('[data-preview-content="true"]') : document;
          element = root?.querySelector(hint.selector) ?? null;
        }
      } catch {
        element = null;
      }

      const width = Math.min(358, window.innerWidth - 32);
      const height = 200;

      if (!element) {
        setHighlight(null);

        if (preview) {
          setPosition({
            left: Math.max(16, (window.innerWidth - width) / 2),
            top: Math.max(16, (window.innerHeight - height) / 2),
          });
        } else {
          setPosition(null);
        }
        return;
      }

      const rect = element.getBoundingClientRect();

      if (preview) setHighlight({ top: rect.top, left: rect.left, width: rect.width, height: rect.height });
      else setHighlight(null);

      const centerX = rect.left + rect.width / 2 - width / 2;
      const centerY = rect.top + rect.height / 2 - height / 2;
      let left = centerX;
      let top = rect.bottom + 12;

      switch (hint.placement) {
        case "top":
          top = rect.top - height - 12;
          break;

        case "bottom":
          top = rect.bottom + 12;
          break;

        case "left":
          left = rect.left - width - 12;
          top = centerY;
          break;

        case "right":
          left = rect.right + 12;
          top = centerY;
          break;

        case "left-top":
          left = rect.left - width - 12;
          top = rect.top;
          break;

        case "right-top":
          left = rect.right + 12;
          top = rect.top;
          break;

        case "center":
          left = centerX;
          top = centerY;
          break;
      }

      setPosition({
        left: Math.max(16, Math.min(left, window.innerWidth - width - 16)),
        top: Math.max(16, Math.min(top, window.innerHeight - height - 16)),
      });
    };
    update();

    const handleTargetClick = (event: MouseEvent) => {
      if (!preview || busy || !hint.selector) return;

      const clickedElement = event.target;
      if (!(clickedElement instanceof Element)) return;
      if (clickedElement.closest(`.${styles.hint}`) || clickedElement.closest('[data-preview-control="true"]')) return;

      let target: Element | null;

      try {
        target = clickedElement.closest(hint.selector);
      } catch {
        return;
      }
      if (!target) return;

      const previewContent = document.querySelector('[data-preview-content="true"]');
      if (!previewContent || !previewContent.contains(clickedElement) || !previewContent.contains(target)) return;
  
      event.preventDefault();
      event.stopPropagation();
      event.stopImmediatePropagation();
      onNextRef.current();
    };
    if (preview) document.addEventListener("click", handleTargetClick, true);

    const observer = new MutationObserver(update);
    observer.observe(document.body, { childList: true, subtree: true });

    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);

    return () => {
      observer.disconnect();
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
      if (preview) document.removeEventListener("click", handleTargetClick, true);
    };
  }, [hint.selector, hint.placement, preview, busy]);

  if (!position) return null;

  return (
    <>
      {preview && (
        <>
          {highlight ? (
            <div
              className={styles.spotlight}
              style={{top: highlight.top, left: highlight.left, width: highlight.width, height: highlight.height}}
            />
          ) : (<div className={styles.backdrop} />)}
        </>
      )}

      <aside className={styles.hint} style={{top: position.top, left: position.left}}>
        <button className={styles.hint__close} type="button" onClick={onClose} disabled={busy}>×</button>
        <div className={styles.hint__header}>
          <h3 className={styles.hint__header__title}>{hint.title}</h3>
          <p className={styles.hint__header__count}>{index + 1}/{total}</p>
        </div>
        <p className={styles.hint__content}>{hint.content}</p>

        {preview && !highlight && (
          <p className={styles.hint__warning}>Элемент не найден: <code>{hint.selector}</code></p>
        )}

        <button type="button" onClick={onNext} disabled={busy}>{index === total - 1 ? "Завершить" : "Далее"}</button>
      </aside>
    </>
  );
}
