import { memo, useState } from "react";
import styles from "./Styles.module.scss";

interface SelectOption {
  value: string;
  label: string;
}

interface SelectProps {
  value: string;
  options: SelectOption[];
  onChange: (value: string) => void;
  className?: string;
}

function SelectComponent({ value, options, onChange, className = "" }: SelectProps) {
  const [isOpen, setIsOpen] = useState(false);
  const selectedOption = options.find((option) => option.value === value);

  return (
    <div className={`${styles.select} ${className}`}>
      <button type="button" className={styles.select__control} onClick={() => setIsOpen((prev) => !prev)}>
        <span>{selectedOption?.label}</span>
        <span className={styles.select__arrow} />
      </button>

      {isOpen && (
        <div className={styles.select__menu}>
          {options.map((option) => (
            <button
              key={option.value}
              type="button"
              className={styles.select__option}
              onClick={() => {
                onChange(option.value);
                setIsOpen(false);
              }}
            >
              {option.label}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export const Select = memo(SelectComponent);
