export type HintPlacement = "top" | "bottom" | "center" | "left-top" | "right-top" | "left" | "right";

export interface Hint {
  id: string;
  tour_version_id: string;
  step: number;
  title: string;
  content: string;
  selector: string;
  placement: HintPlacement;
  created_at: string;
  updated_at: string;
}

export interface CreateHintRequest {
  title: string;
  content: string;
  selector: string;
  placement: HintPlacement;
}

export type HintPatchRequest = Partial<CreateHintRequest>;