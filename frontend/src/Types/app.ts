export interface AppTheme {
  primary_color: string;
  background_color: string;
  text_color: string;
}

export type AppThemePatch = Partial<AppTheme>;

export interface App {
  id: string;
  name: string;
  public_key: string;
  allowed_origins: string[];
  theme: AppTheme;
}

export const DEFAULT_APP_THEME: AppTheme = {
  primary_color: "#696CF1",
  background_color: "#FFFFFF",
  text_color: "#181818",
};

export const HEX_COLOR_PATTERN = /^#[\da-fA-F]{6}$/;
