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

export async function fetchWeatherProducts() {
  const response = await fetch(`${API_BASE_URL}/weather/products`);
  if (!response.ok) {
    throw new Error('Failed to fetch weather products');
  }
  return response.json();
}

export async function fetchProductImages(product: string, date?: Date) {
  const params = new URLSearchParams();
  let formattedDate: string | undefined;
  
  if (date) {
    formattedDate = format(toZonedTime(date, TIME_ZONE), 'yyyy-MM-dd');
    params.append('date', formattedDate);
  }
  
  const response = await fetch(`${API_BASE_URL}/weather/products/${product}?${params}`);
  if (!response.ok) {
    throw new Error(`Failed to fetch images for product: ${product}`);
  }
  const data = await response.json();
  
  // Workaround for API UTC/CST date mismatch bug
  // If no images found and we have a date, try adjacent dates
  if (date && formattedDate && (!data.images || data.images === null || data.images.length === 0)) {
    console.log(`No images for ${formattedDate}, trying adjacent dates...`);
    
    // Try previous day
    const prevDate = new Date(date);
    prevDate.setDate(prevDate.getDate() - 1);
    const prevFormattedDate = format(toZonedTime(prevDate, TIME_ZONE), 'yyyy-MM-dd');
    const prevParams = new URLSearchParams();
    prevParams.append('date', prevFormattedDate);
    
    try {
      const prevResponse = await fetch(`${API_BASE_URL}/weather/products/${product}?${prevParams}`);
      if (prevResponse.ok) {
        const prevData = await prevResponse.json();
        if (prevData.images && prevData.images.length > 0) {
          console.log(`Found images on previous date: ${prevFormattedDate}`);
          return prevData;
        }
      }
    } catch (e) {
      console.error('Error fetching previous date:', e);
    }
    
    // Try next day
    const nextDate = new Date(date);
    nextDate.setDate(nextDate.getDate() + 1);
    const nextFormattedDate = format(toZonedTime(nextDate, TIME_ZONE), 'yyyy-MM-dd');
    const nextParams = new URLSearchParams();
    nextParams.append('date', nextFormattedDate);
    
    try {
      const nextResponse = await fetch(`${API_BASE_URL}/weather/products/${product}?${nextParams}`);
      if (nextResponse.ok) {
        const nextData = await nextResponse.json();
        if (nextData.images && nextData.images.length > 0) {
          console.log(`Found images on next date: ${nextFormattedDate}`);
          return nextData;
        }
      }
    } catch (e) {
      console.error('Error fetching next date:', e);
    }
  }
  
  return data;
}

export async function fetchEmwinCategories() {
  const response = await fetch(`${API_BASE_URL}/emwin/text/categories`);
  if (!response.ok) {
    throw new Error('Failed to fetch EMWIN categories');
  }
  return response.json();
}

export async function fetchEmwinFiles(category: string, date?: Date, station?: string, office?: string) {
  const params = new URLSearchParams();
  params.append('category', category);
  
  if (date) {
    const formattedDate = format(toZonedTime(date, TIME_ZONE), 'yyyy-MM-dd');
    params.append('date', formattedDate);
  }
  
  if (station) {
    params.append('station', station);
  }
  
  if (office) {
    params.append('office', office);
  }
  
  const response = await fetch(`${API_BASE_URL}/emwin/text/files?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch EMWIN files');
  }
  return response.json();
}

export async function fetchWeatherOffices() {
  const response = await fetch(`${API_BASE_URL}/weather/offices`);
  if (!response.ok) {
    throw new Error('Failed to fetch weather offices');
  }
  return response.json();
}

export async function fetchEmwinContent(key: string) {
  const params = new URLSearchParams();
  params.append('key', key);
  
  const response = await fetch(`${API_BASE_URL}/emwin/text/content?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch EMWIN content');
  }
  return response.json();
}

