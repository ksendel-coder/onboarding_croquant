import { memo } from "react";
import cn from "classnames";
import styles from "./Styles.module.scss";
import { Button } from "../Button";

type CheckboxProps = {
  checked: boolean;
  onChange: (checked: boolean) => void;
  label?: string;
  className?: string;
};

function CheckboxComponent({ checked, onChange, label, className }: CheckboxProps) {
  return (
    <label className={cn(styles.checkbox, className)}>
      <Button
        type="button"
        aria-checked={checked}
        className={styles.checkbox__box}
        onClick={() => onChange(!checked)}
      >
        {checked && (
          <svg width="10" height="8" viewBox="0 0 10 8" fill="none">
            <path d="M0.75 3.58L3.58 6.41L9.25 0.75" stroke="#0D99FF" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        )}
      </Button>
      {label && (<span className={cn(styles.checkbox__label, checked && styles.checkbox__label__active)}>{label}</span>)}
    </label>
  );
}

export const Checkbox = memo(CheckboxComponent);