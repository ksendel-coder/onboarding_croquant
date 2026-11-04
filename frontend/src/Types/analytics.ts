export interface StepAnalytics {
  hint_id: string;
  step: number;
  shown: number;
  completed: number;
  drop_off: number;
}
export interface TourAnalytics {
  tour_id: string;
  tour_version_id: string;
  started: number;
  completed: number;
  dismissed: number;
  completion_rate: number; 
  steps: StepAnalytics[];
}