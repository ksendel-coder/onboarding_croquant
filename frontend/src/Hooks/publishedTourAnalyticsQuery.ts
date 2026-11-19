import { queryOptions } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function publishedTourAnalyticsQuery(tourId: string) {
  return queryOptions({
    queryKey: ["published-tour-analytics", tourId],
    queryFn: async () => {
      const versions = await onboardingAPI.getTourVersions(tourId);
      const version = versions.find((item) => item.status === "published");
      if (!version) return null;

      const analytics = await onboardingAPI.getTourAnalytics(tourId, version.id);
      return { version, analytics };
    },
  });
}
