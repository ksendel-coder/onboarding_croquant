const SUBJECT_KEY = "croquant_subject_id";
const ATTEMPT_PREFIX = "croquant_attempt_";
const TARGET_TIMEOUT_MS = 4_000;
const DEFAULT_THEME = Object.freeze({
  primary_color: "#696CF1",
  background_color: "#FFFFFF",
  text_color: "#181818",
});

function setWidgetTheme(host, theme) {
  const colors = Object.fromEntries(
    Object.entries(DEFAULT_THEME).map(([key, fallback]) => {
      const value = theme?.[key];
      return [key, typeof value === "string" && /^#[0-9a-f]{6}$/i.test(value) ? value.toUpperCase() : fallback];
    }),
  );
  const rgb = [1, 3, 5].map((index) => Number.parseInt(colors.primary_color.slice(index, index + 2), 16) / 255);
  const [red, green, blue] = rgb.map((channel) =>
    channel <= 0.04045 ? channel / 12.92 : ((channel + 0.055) / 1.055) ** 2.4,
  );
  const luminance = 0.2126 * red + 0.7152 * green + 0.0722 * blue;

  host.style.setProperty("--croquant-primary", colors.primary_color);
  host.style.setProperty("--croquant-background", colors.background_color);
  host.style.setProperty("--croquant-text", colors.text_color);
  host.style.setProperty("--croquant-button-text", luminance > 0.179 ? "#000000" : "#FFFFFF");
}

function getSubjectId() {
  let id = localStorage.getItem(SUBJECT_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(SUBJECT_KEY, id);
  }
  return id;
}

function getSessionId() {
  const key = `${ATTEMPT_PREFIX}runtime_session`;
  let id = sessionStorage.getItem(key);
  if (!id) {
    id = crypto.randomUUID();
    sessionStorage.setItem(key, id);
  }
  return id;
}

function getAttemptId({ attempt_id, tour, tour_version_id }) {
  if (attempt_id) return attempt_id;
  const tourId = tour.id;
  const versionId = tour_version_id;
  const key = `${ATTEMPT_PREFIX}${tourId}_${versionId}`;
  let id = localStorage.getItem(key);
  if (!id) {
    id = sessionStorage.getItem(key) ?? crypto.randomUUID();
    localStorage.setItem(key, id);
  }
  return id;
}

function endpoint(apiUrl, path) {
  return `${apiUrl.replace(/\/$/, "")}/${path}`;
}

async function resolveTour(apiUrl, appKey, signal) {
  const response = await fetch(endpoint(apiUrl, "resolve"), {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-App-Key": appKey },
    signal,
    body: JSON.stringify({
      url: window.location.href,
      subject_id: getSubjectId(),
      session_id: getSessionId(),
      props: {},
    }),
  });

  if (response.status === 204) return null;
  if (!response.ok) throw new Error(`Resolve failed: ${response.status}`);
  return response.json();
}

async function sendEvent(apiUrl, appKey, event, signal) {
  const body = JSON.stringify(event);

  for (let attempt = 0; attempt < 3; attempt++) {
    signal.throwIfAborted();
    try {
      const response = await fetch(endpoint(apiUrl, "events"), {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-App-Key": appKey },
        body,
        keepalive: true,
      });

      if (response.ok) return;
      if (response.status !== 429 && response.status < 500) throw new Error(`Event rejected: ${response.status}`);
      if (attempt === 2) throw new Error(`Event failed after retries: ${response.status}`);

    } catch (error) {
      if (error instanceof Error && error.message.startsWith("Event rejected:")) throw error;
      if (attempt === 2) throw error;
    }
    await new Promise((resolve) => setTimeout(resolve, 400 * (attempt + 1)));
  }
}

function makeEvent(result, type, hintId) {
  const attempt = getAttemptId(result);
  const identity = `croquant_event_${attempt}_${type}_${hintId ?? "tour"}`;

  let eventId = sessionStorage.getItem(identity);
  if (!eventId) {
    eventId = crypto.randomUUID();
    sessionStorage.setItem(identity, eventId);
  }
  return {
    event_id: eventId,
    attempt_id: attempt,
    subject_id: getSubjectId(),
    tour_id: result.tour.id,
    tour_version_id: result.tour_version_id,
    type,
    ...(hintId ? { hint_id: hintId } : {}),
  };
}

const styles = `
  :host {
    all: initial;
    position: fixed;
    z-index: 10000;
    font: 16px system-ui, sans-serif;
    color: var(--croquant-text, #181818);
  }

  aside {
    position: fixed;
    width: min(358px, calc(100vw - 32px));
    max-height: calc(100vh - 32px);
    overflow: auto;
    padding: 24px;
    border: 3px solid var(--croquant-primary, #696cf1);
    border-radius: 20px;
    background: var(--croquant-background, #fff);
    color: var(--croquant-text, #181818);
    box-sizing: border-box;
    box-shadow: 0 4px 20px #17233b24;
  }

  .spotlight {
    position: fixed;
    outline: 2px solid #fff;
    box-shadow: 0 0 0 200vmax #0009;
    pointer-events: none;
  }

  header {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 12px;
    padding-right: 22px;
  }

  h3, p { margin: 0; overflow-wrap: anywhere; }
  h3 { font-size: 22px; }
  header span { color: var(--croquant-primary, #696cf1); white-space: nowrap; }
  p { margin: 12px 0; color: var(--croquant-text, #181818); }

  button {
    border: 0;
    border-radius: 8px;
    padding: 7px 14px;
    background: var(--croquant-primary, #696cf1);
    color: var(--croquant-button-text, #fff);
    font: inherit;
    cursor: pointer;
  }

  button:disabled { opacity: .55; cursor: default; }
  .next { display: block; margin-left: auto; }
  .close {
    position: absolute;
    top: 4px;
    right: 8px;
    padding: 0 6px;
    background: transparent;
    color: inherit;
    font-size: 26px;
  }

  .error { left: 16px; bottom: 16px; border-color: #fecaca; }
  .error { background: #fff; }
  .error p { color: #991b1b; }
  .error button { background: #5054c9; color: #fff; }
`;

function createElement(tag, text, className) {
  const element = document.createElement(tag);
  if (text) element.textContent = text;
  if (className) element.className = className;
  if (tag === "button") element.type = "button";
  return element;
}

function findTarget(selector) {
  try {
    return selector ? document.querySelector(selector) : null;
  } catch {
    return null;
  }
}

function getStepDestination(hints, index) {
  const path = hints.slice(0, index + 1).findLast((hint) => hint.page_path)?.page_path;
  if (!path?.startsWith("/") || path.startsWith("//")) return null;

  const current = new URL(window.location.href);
  const destination = new URL(path, current);
  if (destination.origin !== current.origin || destination.search || destination.hash) return null;

  const comparablePath = (url) => {
    try {
      return decodeURIComponent(url.pathname).replace(/\/+$/, "");
    } catch {
      return url.pathname;
    }
  };
  return comparablePath(current) === comparablePath(destination) ? null : destination;
}

function getTargetRect(target) {
  const rect = target?.getBoundingClientRect();
  if (!rect || rect.width <= 0 || rect.height <= 0 ||
    target.checkVisibility?.({ opacityProperty: true, visibilityProperty: true }) === false) return null;
  return rect;
}

function isInViewport(rect) {
  return rect && rect.right > 0 && rect.bottom > 0 &&
    rect.left < window.innerWidth && rect.top < window.innerHeight;
}

function positionHint(panel, spotlight, hint, rect) {
  const visible = isInViewport(rect);
  panel.hidden = spotlight.hidden = !visible;
  if (!visible) return false;

  Object.assign(spotlight.style, {
    left: `${rect.left}px`,
    top: `${rect.top}px`,
    width: `${rect.width}px`,
    height: `${rect.height}px`,
  });

  const { width, height } = panel.getBoundingClientRect();
  const centerX = rect.left + (rect.width - width) / 2;
  const centerY = rect.top + (rect.height - height) / 2;
  const placements = {
    top: [centerX, rect.top - height - 12],
    bottom: [centerX, rect.bottom + 12],
    left: [rect.left - width - 12, centerY],
    right: [rect.right + 12, centerY],
    "left-top": [rect.left - width - 12, rect.top],
    "right-top": [rect.right + 12, rect.top],
    center: [centerX, centerY],
  };
  const [left, top] = Object.hasOwn(placements, hint.placement)
    ? placements[hint.placement]
    : placements.bottom;

  panel.style.left = `${Math.max(16, Math.min(left, window.innerWidth - width - 16))}px`;
  panel.style.top = `${Math.max(16, Math.min(top, window.innerHeight - height - 16))}px`;
  return true;
}

export function mountOnboardingWidget({ appKey, apiUrl }) {
  const controller = new AbortController();
  const { signal } = controller;
  const host = createElement("croquant-onboarding");
  const shadow = host.attachShadow({ mode: "open" });
  const content = createElement("div");
  shadow.append(createElement("style", styles), content);
  document.body.append(host);

  let result;
  let hints = [];
  let index = 0;
  let busy = false;
  let paintFrame = null;
  let targetTimeout = null;
  let updatePosition = () => {};
  const observer = new MutationObserver(() => updatePosition());
  observer.observe(document.body, { attributes: true, childList: true, subtree: true });
  window.addEventListener("scroll", reposition, { capture: true, signal });
  window.addEventListener("resize", reposition, { signal });

  function reposition() {
    updatePosition();
  }

  function destroy() {
    controller.abort();
    observer.disconnect();
    cancelPaint();
    cancelTargetTimeout();
    host.remove();
  }

  function cancelPaint() {
    if (paintFrame !== null) window.cancelAnimationFrame(paintFrame);
    paintFrame = null;
  }

  function cancelTargetTimeout() {
    clearTimeout(targetTimeout);
    targetTimeout = null;
  }

  function post(type, hintId) {
    return sendEvent(apiUrl, appKey, makeEvent(result, type, hintId), signal);
  }

  async function run(
    action,
    failureMessage,
    onSuccess = renderHint,
    retryAction = () => void run(action, failureMessage, onSuccess),
  ) {
    if (busy || signal.aborted) return;
    busy = true;
    content.querySelectorAll("button").forEach((button) => { button.disabled = true; });
    try {
      await action();
    } catch {
      if (signal.aborted) return;
      cancelPaint();
      cancelTargetTimeout();
      updatePosition = () => {};
      const panel = createElement("aside", "", "error");
      panel.setAttribute("role", "alert");
      const retry = createElement("button", "Повторить");
      retry.addEventListener("click", retryAction);
      panel.append(createElement("p", failureMessage), retry);
      content.replaceChildren(panel);
      return;
    } finally {
      busy = false;
    }
    if (!signal.aborted) onSuccess();
  }

  async function advance(type = "hint_completed") {
    await post(type, hints[index].id);
    if (index === hints.length - 1) {
      await post("tour_completed");
      destroy();
      return;
    }

    index++;
  }

  async function dismiss() {
    await post("tour_dismissed");
    destroy();
  }

  function renderHint() {
    cancelTargetTimeout();
    const destination = getStepDestination(hints, index);
    if (destination) {
      window.location.assign(destination.href);
      destroy();
      return;
    }

    const hint = hints[index];
    const spotlight = createElement("div", "", "spotlight");
    spotlight.setAttribute("aria-hidden", "true");
    const panel = createElement("aside");
    panel.setAttribute("aria-label", "Подсказка онбординга");

    const close = createElement("button", "×", "close");
    close.setAttribute("aria-label", "Закрыть подсказку");
    close.addEventListener("click", () => void run(dismiss, "Не удалось закрыть сценарий."));

    const next = createElement("button", index === hints.length - 1 ? "Завершить" : "Далее", "next");
    next.addEventListener("click", () => void run(advance, "Не удалось перейти к следующему шагу."));

    const header = createElement("header");
    header.append(createElement("h3", hint.title), createElement("span", `${index + 1}/${hints.length}`));
    close.disabled = next.disabled = true;
    panel.append(close, header, createElement("p", hint.content), next);
    content.replaceChildren(spotlight, panel);

    let shown = false;
    let scrolledTarget;
    updatePosition = () => {
      const target = findTarget(hint.selector);
      let rect = getTargetRect(target);
      if (rect && target !== scrolledTarget) {
        scrolledTarget = target;
        target.scrollIntoView({
          block: isInViewport(rect) ? "nearest" : "center",
          inline: "nearest",
          behavior: window.matchMedia("(prefers-reduced-motion: reduce)").matches ? "instant" : "smooth",
        });
        rect = getTargetRect(target);
      }
      if (!positionHint(panel, spotlight, hint, rect)) {
        cancelPaint();
        if (targetTimeout === null && !busy) {
          targetTimeout = setTimeout(() => {
            targetTimeout = null;
            if (isInViewport(getTargetRect(findTarget(hint.selector)))) {
              reposition();
              return;
            }
            void run(() => advance("hint_skipped"), "Не удалось пропустить шаг.");
          }, TARGET_TIMEOUT_MS);
        }
        return;
      }
      cancelTargetTimeout();
      if (shown || busy || paintFrame !== null) return;

      // Leave a paint between revealing the panel and recording its impression.
      paintFrame = window.requestAnimationFrame(() => {
        paintFrame = window.requestAnimationFrame(() => {
          paintFrame = null;
          if (signal.aborted) return;
          if (!positionHint(panel, spotlight, hint, getTargetRect(findTarget(hint.selector)))) {
            reposition();
            return;
          }
          void run(
            () => post("hint_shown", hint.id),
            "Не удалось записать показ подсказки.",
            () => {
              shown = true;
              close.disabled = next.disabled = false;
              reposition();
            },
            renderHint,
          );
        });
      });
    };
    reposition();
  }

  async function load() {
    if (!result) {
      result = await resolveTour(apiUrl, appKey, signal);
      signal.throwIfAborted();
      if (result) setWidgetTheme(host, result.theme);
      hints = [...(result?.tour.hints ?? [])].sort((a, b) => a.step - b.step);
      if (!hints.length) {
        destroy();
        return;
      }
      index = Math.max(0, hints.findIndex((hint) => hint.id === result.current_hint_id));
    }

    if (!result.current_hint_id) {
      await post("tour_completed");
      destroy();
      return;
    }

    await post("tour_started");
  }

  void run(load, "Не удалось загрузить сценарий.");
  return destroy;
}
