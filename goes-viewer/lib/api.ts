import { parseISO, format } from 'date-fns';
import { toZonedTime } from 'date-fns-tz';

const API_BASE_URL = '/api/proxy';
const TIME_ZONE = 'America/Chicago'; // US Central Time

export async function fetchLatestImage() {
  const response = await fetch(`${API_BASE_URL}/latest`);
  if (!response.ok) {
    throw new Error('Failed to fetch latest image');
  }
  return response.json();
}

export async function fetchAvailableDates() {
  const response = await fetch(`${API_BASE_URL}/available-dates`);
  if (!response.ok) {
    throw new Error('Failed to fetch available dates');
  }
  const data = await response.json();
  return {
    availableDates: data.availableDates.map((dateString: string) => {
      const date = toZonedTime(parseISO(dateString), TIME_ZONE);
      return new Date(date.getFullYear(), date.getMonth(), date.getDate());
    })
  };
}

export async function fetchImagesByDate(date: Date) {
  const formattedDate = format(toZonedTime(date, TIME_ZONE), 'yyyy-MM-dd');
  const response = await fetch(`${API_BASE_URL}/archive/${formattedDate}`);
  if (!response.ok) {
    throw new Error('Failed to fetch images by date');
  }
  return response.json();
}

