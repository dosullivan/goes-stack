import { NextRequest, NextResponse } from 'next/server'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000'

export async function GET(request: NextRequest) {
  try {
    const searchParams = request.nextUrl.searchParams
    const key = searchParams.get('key')
    
    if (!key) {
      return NextResponse.json({ error: 'Key parameter is required' }, { status: 400 })
    }
    
    const url = `${API_BASE_URL}/emwin/text/content?key=${encodeURIComponent(key)}`
    const response = await fetch(url)
    
    if (!response.ok) {
      throw new Error(`API responded with status: ${response.status}`)
    }
    
    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching EMWIN content:', error)
    return NextResponse.json({ content: '' })
  }
}