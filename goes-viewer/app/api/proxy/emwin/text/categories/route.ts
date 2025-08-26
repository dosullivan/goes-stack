import { NextResponse } from 'next/server'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000'

export async function GET() {
  try {
    const response = await fetch(`${API_BASE_URL}/emwin/text/categories`)
    
    if (!response.ok) {
      throw new Error(`API responded with status: ${response.status}`)
    }
    
    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching EMWIN categories:', error)
    // Return mock data for development
    return NextResponse.json({
      categories: [
        'Forecasts',
        'Warnings',
        'Observations',
        'Marine',
        'Aviation'
      ]
    })
  }
}