"use client";

import { useEffect, useId, useRef, useState } from "react";

export interface SelectOption {
  value: string | number;
  label: string;
}

interface SelectProps {
  value: string | number | "";
  onChange: (value: string | number | "") => void;
  options: SelectOption[];
  placeholder?: string;
  disabled?: boolean;
  id?: string;
}

export default function Select({
  value,
  onChange,
  options,
  placeholder = "请选择",
  disabled = false,
  id,
}: SelectProps) {
  const autoId = useId();
  const selectId = id ?? autoId;
  const listboxId = `${selectId}-listbox`;
  const containerRef = useRef<HTMLDivElement>(null);
  const [open, setOpen] = useState(false);

  const selected = options.find((opt) => String(opt.value) === String(value));

  useEffect(() => {
    if (!open) return;

    function handlePointerDown(e: MouseEvent) {
      if (
        containerRef.current &&
        !containerRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    }

    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }

    document.addEventListener("mousedown", handlePointerDown);
    document.addEventListener("keydown", handleKeyDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [open]);

  function choose(next: string | number) {
    onChange(next);
    setOpen(false);
  }

  return (
    <div ref={containerRef} className="relative mt-2">
      <button
        id={selectId}
        type="button"
        disabled={disabled}
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={listboxId}
        onClick={() => !disabled && setOpen((prev) => !prev)}
        className="flex w-full items-center justify-between rounded-xl border border-stone-300 bg-white px-4 py-3 text-left text-sm text-stone-900 outline-none transition hover:border-stone-400 focus:border-amber-600 disabled:cursor-not-allowed disabled:opacity-60"
      >
        <span className={selected ? "text-stone-900" : "text-stone-400"}>
          {selected?.label ?? placeholder}
        </span>
        <svg
          className={`h-4 w-4 shrink-0 text-stone-500 transition ${open ? "rotate-180" : ""}`}
          viewBox="0 0 20 20"
          fill="currentColor"
          aria-hidden
        >
          <path
            fillRule="evenodd"
            d="M5.23 7.21a.75.75 0 011.06.02L10 10.94l3.71-3.71a.75.75 0 111.06 1.06l-4.24 4.25a.75.75 0 01-1.06 0L5.21 8.29a.75.75 0 01.02-1.08z"
            clipRule="evenodd"
          />
        </svg>
      </button>

      {open && (
        <ul
          id={listboxId}
          role="listbox"
          aria-labelledby={selectId}
          className="absolute z-20 mt-1 max-h-60 w-full overflow-auto rounded-xl border border-stone-200 bg-white py-1 shadow-lg"
        >
          {options.map((opt) => {
            const isSelected = String(opt.value) === String(value);
            return (
              <li key={opt.value} role="option" aria-selected={isSelected}>
                <button
                  type="button"
                  onClick={() => choose(opt.value)}
                  className={`flex w-full px-4 py-2.5 text-left text-sm transition ${
                    isSelected
                      ? "bg-amber-50 font-medium text-amber-900"
                      : "text-stone-800 hover:bg-stone-50"
                  }`}
                >
                  {opt.label}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
