import type { ResolveResult, WidgetEvent } from "../../Types";
import { SUBJECT_KEY, ATTEMPT_PREFIX } from "../../Utils/constants";

export function getSubjectId(): string {
  let id = localStorage.getItem(SUBJECT_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(SUBJECT_KEY, id);
  }
  return id;
}

export function getAttemptId(tourId: string, versionId: string): string {
  const key = `${ATTEMPT_PREFIX}${tourId}_${versionId}`;
  let id = sessionStorage.getItem(key);
  if (!id) {
    id = crypto.randomUUID();
    sessionStorage.setItem(key, id);
  }
  return id;
}

function endpoint(apiUrl: string, path: string): string {
  return `${apiUrl.replace(/\/$/, "")}/${path}`;
}

export async function resolveTour(apiUrl: string, appKey: string): Promise<ResolveResult | null> {
  const response = await fetch(endpoint(apiUrl, "resolve"), {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-App-Key": appKey },
    body: JSON.stringify({
      url: window.location.href,
      subject_id: getSubjectId(),
      session_id: getAttemptId("runtime", "session"),
      props: {},
    }),
  });

  if (response.status === 204) return null;
  if (!response.ok) throw new Error(`Resolve failed: ${response.status}`);

  return (await response.json()) as ResolveResult;
}

export async function sendEvent(apiUrl: string, appKey: string, event: WidgetEvent): Promise<void> {
  const body = JSON.stringify(event);
  
  for (let i = 0; i < 3; i++) {
    try {
      const response = await fetch(endpoint(apiUrl, "events"), {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-App-Key": appKey },
        body,
        keepalive: true,
      });

      if (response.ok) return; 
      if (response.status !== 429 && response.status < 500) throw new Error(`Event rejected: ${response.status}`);

    } catch (error) {
      if (i === 2) throw error;
    }
    await new Promise<void>((resolve) => setTimeout(resolve, 400 * (i + 1)));
  }
  throw new Error("Could not send onboarding event");
}

export function makeEvent(result: ResolveResult, type: WidgetEvent["type"], hintId?: string): WidgetEvent {
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