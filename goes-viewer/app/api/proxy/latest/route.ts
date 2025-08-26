import { NextResponse } from 'next/server';

const API_BASE_URL = process.env.API_BASE_URL || 'http://192.168.7.8:3010';

export async function GET() {
  try {
    const response = await fetch(`${API_BASE_URL}/latest`);
    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error fetching latest image:', error);
    return NextResponse.json({ error: 'Failed to fetch latest image' }, { status: 500 });
  }
}

