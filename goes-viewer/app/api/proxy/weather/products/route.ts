import { NextResponse } from 'next/server'

const API_BASE_URL = process.env.API_BASE_URL || 'http://localhost:3000'

export async function GET() {
  try {
    const response = await fetch(`${API_BASE_URL}/weather/products`)
    
    if (!response.ok) {
      throw new Error(`API responded with status: ${response.status}`)
    }
    
    const data = await response.json()
    return NextResponse.json(data)
  } catch (error) {
    console.error('Error fetching weather products:', error)
    // Return mock data for development
    return NextResponse.json({
      products: [
        // GOES-19 Full Disk products
        { id: 'goes19_fd_fc', name: 'False Color', satellite: 'goes19', region: 'fd', channel: 'fc' },
        { id: 'goes19_fd_ch02', name: 'Channel 02', satellite: 'goes19', region: 'fd', channel: 'ch02' },
        { id: 'goes19_fd_ch07', name: 'Channel 07', satellite: 'goes19', region: 'fd', channel: 'ch07' },
        { id: 'goes19_fd_ch08', name: 'Channel 08', satellite: 'goes19', region: 'fd', channel: 'ch08' },
        { id: 'goes19_fd_ch09', name: 'Channel 09', satellite: 'goes19', region: 'fd', channel: 'ch09' },
        { id: 'goes19_fd_ch13', name: 'Channel 13', satellite: 'goes19', region: 'fd', channel: 'ch13' },
        { id: 'goes19_fd_ch14', name: 'Channel 14', satellite: 'goes19', region: 'fd', channel: 'ch14' },
        { id: 'goes19_fd_ch15', name: 'Channel 15', satellite: 'goes19', region: 'fd', channel: 'ch15' },
        // Enhanced versions
        { id: 'goes19_fd_ch07_enhanced', name: 'Channel 07 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch07_enhanced' },
        { id: 'goes19_fd_ch08_enhanced', name: 'Channel 08 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch08_enhanced' },
        { id: 'goes19_fd_ch09_enhanced', name: 'Channel 09 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch09_enhanced' },
        { id: 'goes19_fd_ch13_enhanced', name: 'Channel 13 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch13_enhanced' },
        { id: 'goes19_fd_ch14_enhanced', name: 'Channel 14 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch14_enhanced' },
        { id: 'goes19_fd_ch15_enhanced', name: 'Channel 15 Enhanced', satellite: 'goes19', region: 'fd', channel: 'ch15_enhanced' },
        // Mesoscale regions
        { id: 'goes19_m1_fc', name: 'False Color', satellite: 'goes19', region: 'm1', channel: 'fc' },
        { id: 'goes19_m2_fc', name: 'False Color', satellite: 'goes19', region: 'm2', channel: 'fc' },
        // Level 2 products
        { id: 'goes19_acha', name: 'Cloud Top Height', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_acht', name: 'Cloud Top Temperature', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_dsi', name: 'Derived Stability Indices', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_tpw', name: 'Total Precipitable Water', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_lst', name: 'Land Surface Temperature', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_sst', name: 'Sea Surface Temperature', satellite: 'goes19', category: 'Level 2' },
        { id: 'goes19_rrqpe', name: 'Rainfall Rate QPE', satellite: 'goes19', category: 'Level 2' }
      ]
    })
  }
}