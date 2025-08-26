import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// takes the filename and extracts a human-readable timestamp from it
export function parseTimestamp(filename: string) {
  const parts = filename.match(/c(\d{4})(\d{3})(\d{2})(\d{2})(\d{2})/);
  if (!parts) return '';

  const year = parts[1];
  const dayOfYear = parseInt(parts[2], 10);
  const hour = parts[3];
  const minute = parts[4];
  const second = parts[5];

  const date = new Date(year);
  date.setUTCDate(dayOfYear);
  date.setUTCHours(
    parseInt(hour, 10),
    parseInt(minute, 10),
    parseInt(second, 10)
  );

  return date.toLocaleString(); // Converts to a readable format
}