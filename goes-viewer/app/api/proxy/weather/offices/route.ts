import { NextResponse } from 'next/server';

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000';

export async function GET() {
  try {
    const response = await fetch(`${API_BASE_URL}/weather/offices`);
    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error fetching weather offices:', error);
    return NextResponse.json({ error: 'Failed to fetch weather offices' }, { status: 500 });
  }
}