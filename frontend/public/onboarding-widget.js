const SUBJECT_KEY = "croquant_subject_id";
const ATTEMPT_PREFIX = "croquant_attempt_";

function getSubjectId() {
  let id = localStorage.getItem(SUBJECT_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(SUBJECT_KEY, id);
  }
  return id;
}

function getAttemptId(tourId, versionId) {
  const key = `${ATTEMPT_PREFIX}${tourId}_${versionId}`;
  let id = sessionStorage.getItem(key);
  if (!id) {
    id = crypto.randomUUID();
    sessionStorage.setItem(key, id);
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
      session_id: getAttemptId("runtime", "session"),
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
  const attempt = getAttemptId(result.tour.id, result.tour_version_id);
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
    color: #181818;
  }

  aside {
    position: fixed;
    width: min(358px, calc(100vw - 32px));
    max-height: calc(100vh - 32px);
    overflow: auto;
    padding: 24px;
    border: 3px solid #696cf1;
    border-radius: 20px;
    background: #fff;
    box-sizing: border-box;
    box-shadow: 0 4px 20px #17233b24;
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
  header span { color: #585bcc; white-space: nowrap; }
  p { margin: 12px 0; color: #4b5563; }

  button {
    border: 0;
    border-radius: 8px;
    padding: 7px 14px;
    background: #5054c9;
    color: #fff;
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
  .error p { color: #991b1b; }
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

function positionHint(panel, hint) {
  const target = findTarget(hint.selector);
  panel.hidden = !target;
  if (!target) return;

  const rect = target.getBoundingClientRect();
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
  let updatePosition = () => {};
  const observer = new MutationObserver(() => updatePosition());
  observer.observe(document.body, { childList: true, subtree: true });
  window.addEventListener("scroll", reposition, { capture: true, signal });
  window.addEventListener("resize", reposition, { signal });

  function reposition() {
    updatePosition();
  }

  function destroy() {
    controller.abort();
    observer.disconnect();
    host.remove();
  }

  function post(type, hintId) {
    return sendEvent(apiUrl, appKey, makeEvent(result, type, hintId), signal);
  }

  async function run(action, failureMessage) {
    if (busy || signal.aborted) return;
    busy = true;
    content.querySelectorAll("button").forEach((button) => { button.disabled = true; });
    try {
      await action();
      if (!signal.aborted) renderHint();
    } catch {
      if (signal.aborted) return;
      updatePosition = () => {};
      const panel = createElement("aside", "", "error");
      panel.setAttribute("role", "alert");
      const retry = createElement("button", "Повторить");
      retry.addEventListener("click", () => void run(action, failureMessage));
      panel.append(createElement("p", failureMessage), retry);
      content.replaceChildren(panel);
    } finally {
      busy = false;
    }
  }

  async function advance() {
    await post("hint_completed", hints[index].id);
    if (index === hints.length - 1) {
      await post("tour_completed");
      destroy();
      return;
    }

    await post("hint_shown", hints[index + 1].id);
    index++;
  }

  async function dismiss() {
    await post("tour_dismissed");
    destroy();
  }

  function renderHint() {
    const hint = hints[index];
    const panel = createElement("aside");
    panel.setAttribute("aria-label", "Подсказка онбординга");

    const close = createElement("button", "×", "close");
    close.setAttribute("aria-label", "Закрыть подсказку");
    close.addEventListener("click", () => void run(dismiss, "Не удалось закрыть сценарий."));

    const next = createElement("button", index === hints.length - 1 ? "Завершить" : "Далее", "next");
    next.addEventListener("click", () => void run(advance, "Не удалось перейти к следующему шагу."));

    const header = createElement("header");
    header.append(createElement("h3", hint.title), createElement("span", `${index + 1}/${hints.length}`));
    panel.append(close, header, createElement("p", hint.content), next);
    content.replaceChildren(panel);

    updatePosition = () => positionHint(panel, hint);
    reposition();
  }

  async function start() {
    await post("tour_started");
    await post("hint_shown", hints[index].id);
  }

  async function load() {
    if (!result) {
      result = await resolveTour(apiUrl, appKey, signal);
      signal.throwIfAborted();
      hints = [...(result?.tour.hints ?? [])].sort((a, b) => a.step - b.step);
      if (!hints.length) {
        destroy();
        return;
      }
      index = Math.max(0, hints.findIndex((hint) => hint.id === result.current_hint_id));
    }

    await start();
  }

  void run(load, "Не удалось загрузить сценарий.");
  return destroy;
}
