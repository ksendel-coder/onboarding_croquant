function uniqueInDocument(doc: Document, selector: string): boolean {
  try {
    return doc.querySelectorAll(selector).length === 1;
  } catch {
    return false;
  }
}

export function selectorForElement(element: Element): string {
  const doc = element.ownerDocument;
  if (element.id) {
    const byId = `#${CSS.escape(element.id)}`;
    if (uniqueInDocument(doc, byId)) return byId;
  }

  const testId = element.getAttribute("data-testid");
  if (testId) {
    const byTestId = `[data-testid="${CSS.escape(testId)}"]`;
    if (uniqueInDocument(doc, byTestId)) return byTestId;
  }

  const parts: string[] = [];
  let current: Element | null = element;

  while (current && current !== doc.documentElement) {
    const parent: Element | null = current.parentElement;
    if (!parent) break;

    const siblings = Array.from(parent.children).filter((child) => child.tagName === current?.tagName);
    parts.unshift(`${current.tagName.toLowerCase()}:nth-of-type(${siblings.indexOf(current) + 1})`);

    const candidate = parts.join(" > ");
    if (uniqueInDocument(doc, candidate)) return candidate;

    current = parent;
  }
  
  return parts.join(" > ") || element.tagName.toLowerCase();
}

export function isValidSelector(selector: string): boolean {
  if (!selector.trim()) return false;
  try {
    document.querySelector(selector);
    return true;
  } catch {
    return false;
  }
}