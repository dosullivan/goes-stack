import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// takes the filename or URL and extracts a human-readable timestamp from it
export function parseTimestamp(input: string | undefined) {
  if (!input) return '';
  
  // Ensure we have a string
  const str = String(input);
  
  // Try new format first: GOES19_FD_CH02_20250826T023020Z
  let parts = str.match(/(\d{8})T(\d{6})Z/);
  if (parts) {
    const dateStr = parts[1]; // YYYYMMDD
    const timeStr = parts[2]; // HHMMSS
    
    const year = dateStr.substring(0, 4);
    const month = dateStr.substring(4, 6);
    const day = dateStr.substring(6, 8);
    const hour = timeStr.substring(0, 2);
    const minute = timeStr.substring(2, 4);
    const second = timeStr.substring(4, 6);
    
    const date = new Date(`${year}-${month}-${day}T${hour}:${minute}:${second}Z`);
    return date.toLocaleString('en-US', { timeZone: 'UTC', hour12: false }) + ' UTC';
  }
  
  // Try old format: c20240360000202
  parts = str.match(/c(\d{4})(\d{3})(\d{2})(\d{2})(\d{2})/);
  if (parts) {
    const year = parts[1];
    const dayOfYear = parseInt(parts[2], 10);
    const hour = parts[3];
    const minute = parts[4];
    const second = parts[5];

    const date = new Date(year + '-01-01');
    date.setUTCDate(dayOfYear);
    date.setUTCHours(
      parseInt(hour, 10),
      parseInt(minute, 10),
      parseInt(second, 10)
    );

    return date.toLocaleString('en-US', { timeZone: 'UTC', hour12: false }) + ' UTC';
  }
  
  return '';
}