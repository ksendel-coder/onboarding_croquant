import { useQuery } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";
import { getCurrentApp } from "../Api/helpers/helpers";

export function useToursQuery() {
  return useQuery({
    queryKey: ["tours"],
    queryFn: async () => {
      const app = await getCurrentApp();
      return onboardingAPI.getTours((app).id);
    },
  });
}