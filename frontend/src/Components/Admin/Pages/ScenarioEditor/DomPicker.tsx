import { useEffect, useRef } from "react";
import { selectorForElement } from "../../../../Utils/domSelector";
import styles from "./DomPicker.module.scss";

interface DomPickerProps {
  active: boolean;
  onSelect: (selector: string) => void;
  onCancel: () => void;
}

export function DomPicker({ active, onSelect, onCancel }: DomPickerProps) {
  const iframeRef = useRef<HTMLIFrameElement>(null);

  useEffect(() => {
    if (!active) return;

    const iframe = iframeRef.current;
    if (!iframe) return;

    let cleanup = () => {};

    const install = () => {
      cleanup();

      let doc: Document | null = null;

      try {
        doc = iframe.contentDocument;
      } catch {
        return;
      }

      if (!doc) return;

      let highlighted: HTMLElement | null = null;
      let oldOutline = "";

      const clearHighlight = () => {
        if (highlighted) {
          highlighted.style.outline = oldOutline;
        }

        highlighted = null;
      };

      const handleHover = (event: MouseEvent) => {
        const target = event.target as HTMLElement;

        if (target.nodeType !== 1) return;

        clearHighlight();

        highlighted = target;
        oldOutline = target.style.outline;

        target.style.outline = "3px solid #706cf5";
      };

      const handleClick = (event: MouseEvent) => {
        event.preventDefault();
        event.stopPropagation();

        const target = event.target as Element;
        if (target.nodeType !== 1) return;

        const selector = selectorForElement(target);

        onSelect(selector);
      };

      doc.addEventListener("mouseover", handleHover, true);
      doc.addEventListener("click", handleClick, true);

      cleanup = () => {
        doc.removeEventListener("mouseover", handleHover, true);
        doc.removeEventListener("click", handleClick, true);
        clearHighlight();
      };
    };

    iframe.addEventListener("load", install);
    install();

    return () => {
      iframe.removeEventListener("load", install);
      cleanup();
    };
  }, [active, onSelect]);

  if (!active) return null;

  return (
    <div className={styles.overlay}>
      <div className={styles.modal}>
        <div className={styles.header}>
          <div>
            <h3>Выбор элемента сайта</h3>
            <p>Нажмите на элемент, для которого нужна подсказка.</p>
          </div>

          <button type="button" onClick={onCancel}>
            ✕
          </button>
        </div>

        <iframe
          ref={iframeRef}
          src="/picker-preview.html"
          title="Выбор DOM-элемента"
          className={styles.preview}
        />
      </div>
    </div>
  );
}
