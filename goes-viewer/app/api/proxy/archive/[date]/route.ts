import { NextRequest, NextResponse } from 'next/server';

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000';

export async function GET(
  request: NextRequest,
  context: { params: Promise<{ date: string }> }
) {
  const { date } = await context.params;

  if (!date) {
    return NextResponse.json({ error: 'Date parameter is missing' }, { status: 400 });
  }

  try {
    const response = await fetch(`${API_BASE_URL}/archive/${date}`);
    
    if (!response.ok) {
      throw new Error(`API responded with status: ${response.status}`);
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error('Error fetching archive data:', error);
    return NextResponse.json(
      { error: 'Failed to fetch archive data', details: (error as Error).message },
      { status: 500 }
    );
  }
}

