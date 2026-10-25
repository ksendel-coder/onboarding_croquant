import { useEffect, useRef, useState } from "react";
import { selectorForElement } from "../../../../Utils/domSelector";

interface DomPickerProps {
  onSelect: (selector: string) => void;
}

export function DomPicker({ onSelect }: DomPickerProps) {
  const ref = useRef<HTMLIFrameElement>(null);
  const [active, setActive] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const iframe = ref.current;
    if (!iframe || !active) return;

    let cleanListeners = () => {};

    function install() {
      cleanListeners();
      let doc: Document | null = null;
      try {
        doc = iframe?.contentDocument ?? null;
      } catch {
        setError("Выбор DOM доступен только для same-origin страницы");
      }
      if (!doc) return;

      let highlighted: HTMLElement | null = null;
      let previousOutline = "";

      const clear = () => {
        if (highlighted) highlighted.style.outline = previousOutline;
        highlighted = null;
      };

      const highlight = (event: Event) => {
        const target = event.target;
        if (!target || (target as Node).nodeType !== 1) return;
        clear();
        const el = target as HTMLElement;
        previousOutline = el.style.outline;
        el.style.outline = "3px solid #7557ea";
        highlighted = el;
      };

      const pick = (event: Event) => {
        event.preventDefault();
        event.stopPropagation();
        const target = event.target;
        if (!target || (target as Node).nodeType !== 1) return;
        onSelect(selectorForElement(target as Element));
        setActive(false);
      };

      doc.addEventListener("mouseover", highlight, true);
      doc.addEventListener("click", pick, true);
      cleanListeners = () => {
        doc?.removeEventListener("mouseover", highlight, true);
        doc?.removeEventListener("click", pick, true);
        clear();
      };
    }

    iframe.addEventListener("load", install);
    install();
    return () => {
      iframe.removeEventListener("load", install);
      cleanListeners();
    };
  }, [active, onSelect]);

  return (
    <section>
      <div className="builderPickerActions">
        <button type="button" onClick={() => {
          setError("");
          setActive((v) => !v);
        }}>
          {active ? "Отменить выбор" : "Выбрать элемент на странице"}
        </button>
        <span>{active ? "Нажмите на элемент внутри страницы" : "Режим просмотра"}</span>
      </div>
      {error && <p role="alert">{error}</p>}
      <iframe
        ref={ref}
        title="Предпросмотр сайта для выбора элемента"
        src="/picker-preview.html"
        className="builderPreview"
      />
    </section>
  );
}