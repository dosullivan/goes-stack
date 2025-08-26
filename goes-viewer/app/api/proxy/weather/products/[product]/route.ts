import { NextRequest, NextResponse } from 'next/server'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000'

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ product: string }> }
) {
  const { product } = await params
  try {
    const searchParams = request.nextUrl.searchParams
    const queryString = searchParams.toString()
    
    const url = `${API_BASE_URL}/weather/products/${product}${queryString ? `?${queryString}` : ''}`
    const response = await fetch(url)
    
    if (!response.ok) {
      throw new Error(`API responded with status: ${response.status}`)
    }
    
    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error(`Error fetching product ${product}:`, error)
    // Return empty array as fallback
    return NextResponse.json({ imageUrls: [] })
  }
}