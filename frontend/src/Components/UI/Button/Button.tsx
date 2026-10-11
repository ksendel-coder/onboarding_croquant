import { memo } from "react";
import styles from "./Styles.module.scss";
import cn from "classnames";

interface ButtonProps {
  size?: "nav";
  color?: "primary" | "transparent" | "grey" | "pink" | "back" | "save" | "hint" | "edit";
  children?: React.ReactNode;
  onClick?: () => void;
  type?: "button" | "submit";
  className?: string;
  disabled?: boolean;
}

function ButtonComponent({ size, color, children, onClick, type = "button", className = "", disabled }: ButtonProps) {
  return (
    <button
      onClick={onClick}
      type={type}
      className={cn(styles.button, styles[`button__${color}`], styles[`button__${size}`], className)}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

export const Button = memo(ButtonComponent);
