import { memo } from "react";
import styles from "./Styles.module.scss";
import { Select } from "../Select";

interface FiltersProps {
  search: string;
  status: string;
  onSearchChange: (value: string) => void;
  onStatusChange: (value: string) => void;
}

function FiltersComponent({ search, status, onSearchChange, onStatusChange }: FiltersProps) {
  return (
    <div className={styles.filters}>
      <label className={styles.filters__search}>
        <svg
          className={styles.filters__search__icon}
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle cx="11" cy="11" r="7" stroke="currentColor" strokeWidth="2" />
          <path
            d="M16.5 16.5L21 21"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
          />
        </svg>

        <input
          type="text"
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          placeholder="Поиск по названию"
        />
      </label>

      <div className={styles.filters__selectWrapper}>
        <Select
          value={status}
          onChange={onStatusChange}
          className={styles.filters__select}
          options={[
            {
              value: "all",
              label: "Все статусы",
            },
            {
              value: "published",
              label: "Опубликован",
            },
            {
              value: "draft",
              label: "Черновик",
            },
          ]}
        />
      </div>
    </div>
  );
}

export const Filters = memo(FiltersComponent);
