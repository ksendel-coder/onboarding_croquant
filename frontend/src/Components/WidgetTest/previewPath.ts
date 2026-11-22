export function widgetTestPreviewPath(path?: string): string | null {
  switch (path?.trim().replace(/\/$/, "")) {
    case undefined:
    case "":
    case "/widget-test":
      return "/widget-test";
    case "/widget-test/profile":
      return "/widget-test/profile";
    default:
      return null;
  }
}
