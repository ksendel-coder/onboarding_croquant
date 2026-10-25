import { useParams } from "react-router-dom";
import { useScenario } from "./useScenario";
import { useScenarioSteps } from "./useScenarioSteps";

export function useScenarioEditor() {
  const { tourId } = useParams<{ tourId?: string }>();
  const steps = useScenarioSteps(tourId);
  const scenarioData = useScenario(tourId, steps.uploadDraftHints);
  return { ...steps, ...scenarioData, busy: steps.busy || scenarioData.saveScenario.isPending };
}
