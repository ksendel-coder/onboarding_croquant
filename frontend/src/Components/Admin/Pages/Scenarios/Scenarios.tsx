import { useEffect, useState } from "react";
import { getCurrentApp } from "../../../../Api/helpers/helpers";
import type { App } from "../../../../Types";

export function Scenarios() {
  const [app, setApp] = useState<App | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getCurrentApp()
      .then(setApp)
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Не удалось загрузить приложение");
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  if (isLoading) return <div>Загрузка...</div>;
  if (error) return <div>{error}</div>;

  return (
    <div>
      <h1>Сценарии</h1>
    </div>
  );
}