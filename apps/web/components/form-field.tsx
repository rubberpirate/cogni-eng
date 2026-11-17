"use client";
import { cloneElement, useId, type ReactElement } from "react";
export function FormField({
  name,
  children,
}: {
  name: string;
  children: ReactElement<{ id?: string }>;
}) {
  const id = useId();
  return (
    <div className="field">
      <label htmlFor={id}>{name}</label>
      {cloneElement(children, { id })}
    </div>
  );
}
