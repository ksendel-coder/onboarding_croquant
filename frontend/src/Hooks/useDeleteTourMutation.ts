import { useMutation, useQueryClient } from "@tanstack/react-query";
import { onboardingAPI } from "../Api/onboarding";

export function useDeleteTourMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (id: string) => {
      await onboardingAPI.deleteTour(id);
      return id;
    },

    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["tours"] }); },
    onError: (error) => { console.error("Ошибка удаления:", error); },
  });
}