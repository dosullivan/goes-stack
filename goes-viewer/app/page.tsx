'use client'

import { useState, useEffect, useCallback, useRef } from 'react'
import Image from 'next/image'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from '@/components/ui/calendar'
import { 
  ChevronLeft, 
  ChevronRight, 
  CalendarIcon, 
  FileQuestion,
  Menu,
  X,
  Download,
  ZoomIn,
  ZoomOut,
  RefreshCw,
  Maximize2,
  Grid3X3,
  SplitSquareHorizontal,
  Play,
  FileText
} from 'lucide-react'
import { 
  fetchLatestImage, 
  fetchAvailableDates, 
  fetchImagesByDate, 
  fetchWeatherProducts,
  fetchProductImages
} from '@/lib/api'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { parseTimestamp, cn } from '@/lib/utils'
import { format } from 'date-fns'
import { ProductSelector } from '@/components/product-selector'
import { ViewModeSelector } from '@/components/view-mode-selector'
import { AnimationControls } from '@/components/animation-controls'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import Link from 'next/link'

type ViewMode = 'single' | 'grid' | 'compare' | 'animation'

interface WeatherProduct {
  id: string
  name: string
  category: string
  satellite?: string
  region?: string
  channel?: string
}

export default function Home() {
  // Original state
  const [currentImage, setCurrentImage] = useState<string | null>(null)
  const [currentDate, setCurrentDate] = useState<Date>(new Date())
  const [availableDates, setAvailableDates] = useState<Date[]>([])
  const [images, setImages] = useState<string[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isImageLoading, setIsImageLoading] = useState(true)
  const [currentTimestamp, setCurrentTimestamp] = useState<string>('')
  
  // Enhanced UI state
  const [viewMode, setViewMode] = useState<ViewMode>('single')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [weatherProducts, setWeatherProducts] = useState<WeatherProduct[]>([])
  const [selectedProduct, setSelectedProduct] = useState<WeatherProduct | null>(null)
  const [isAnimating, setIsAnimating] = useState(false)
  const [animationSpeed, setAnimationSpeed] = useState(500)
  const [compareImages, setCompareImages] = useState<string[]>([])
  const [compareIndices, setCompareIndices] = useState<[number, number]>([0, 1])
  const [zoom, setZoom] = useState(1)
  const [panOffset, setPanOffset] = useState({ x: 0, y: 0 })
  const [isDragging, setIsDragging] = useState(false)
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 })
  const animationRef = useRef<NodeJS.Timeout | null>(null)
  const imageContainerRef = useRef<HTMLDivElement>(null)
  
  // Store zoom/pan settings per product region/type
  const zoomPanSettingsRef = useRef<Record<string, { zoom: number, pan: { x: number, y: number } }>>({})
  const previousProductRef = useRef<WeatherProduct | null>(null)

  // Load initial data and weather products
  useEffect(() => {
    const fetchInitialData = async () => {
      try {
        // Fetch available dates and products
        const [dates, products] = await Promise.all([
          fetchAvailableDates(),
          fetchWeatherProducts()
        ])
        
        setAvailableDates(dates.availableDates)
        
        // Parse and organize products
        const parsedProducts: WeatherProduct[] = products.products ? products.products.map((p: any) => {
          // Extract satellite, region, and channel from key if present
          let satellite, region, channel
          if (p.key) {
            const parts = p.key.split('_')
            if (parts[0] === 'goes19' || parts[0] === 'goeswest') {
              satellite = parts[0]
              region = parts[1]
              channel = parts.slice(2).join('_')
            } else if (p.key.startsWith('fd_') || p.key.startsWith('m1_') || p.key.startsWith('m2_')) {
              satellite = 'goes19'
              const [r, ...ch] = p.key.split('_')
              region = r
              channel = ch.join('_')
            }
          }
          
          return {
            id: p.key || p.id || 'unknown',
            name: p.title || p.name || 'Unknown Product',
            category: p.category || 'general',
            satellite: satellite || p.satellite,
            region: region || p.region,
            channel: channel || p.channel
          }
        }) : []
        
        setWeatherProducts(parsedProducts)
        
        // Set default product - prioritize GOES-19 Full Disk Color
        const defaultProduct = parsedProducts.find(p => 
          p.id === 'fd_fc' || 
          p.id === 'fd_color' || 
          (p.name.toLowerCase().includes('full disk') && p.name.toLowerCase().includes('color')) ||
          (p.satellite === 'goes19' && p.name.toLowerCase().includes('color'))
        ) || parsedProducts.find(p => 
          p.satellite === 'goes19' && p.region === 'fd'
        ) || parsedProducts.find(p => 
          p.satellite === 'goes19'
        ) || parsedProducts[0]
        if (defaultProduct) {
          setSelectedProduct(defaultProduct)
          previousProductRef.current = defaultProduct
        }
        
        if (dates.availableDates.length > 0) {
          const latestDate = dates.availableDates[0]
          setCurrentDate(latestDate)
          // Try to get images for default product on latest date
          if (defaultProduct) {
            try {
              const productImages = await fetchProductImages(defaultProduct.id, latestDate)
              const imageArray = productImages.imageUrls || productImages.images || []
              if (imageArray.length > 0) {
                // Convert image objects to URLs if needed
                const urls = imageArray.map((img: any) => {
                  if (typeof img === 'string') return img
                  if (img.url) return img.url
                  return img
                })
                // Reverse for chronological order (oldest to newest) for animations
                const chronologicalUrls = [...urls].reverse()
                setImages(chronologicalUrls)
                setCurrentIndex(0)
                setCurrentImage(chronologicalUrls[0])
                setCurrentTimestamp(parseTimestamp(chronologicalUrls[0]))
              }
            } catch (e) {
              console.log('No images for default product')
            }
          }
        }
      } catch (error) {
        console.error('Error fetching initial data:', error)
        // Fallback to just available dates
        try {
          const dates = await fetchAvailableDates()
          setAvailableDates(dates.availableDates)
          
          if (dates.availableDates.length > 0) {
            const latestDate = dates.availableDates[0]
            setCurrentDate(latestDate)
          }
        } catch (fallbackError) {
          console.error('Fallback failed:', fallbackError)
        }
      }
    }
    fetchInitialData()
  }, [])

  // Get product group key for zoom/pan persistence
  const getProductGroupKey = (product: WeatherProduct | null): string => {
    if (!product) return 'default'
    
    // Group by region/type for similar products
    if (product.id.includes('fd_') || product.id.includes('Full Disk')) {
      return 'full_disk'
    } else if (product.id.includes('m1_') || product.id.includes('Mesoscale 1')) {
      return 'mesoscale_1'
    } else if (product.id.includes('m2_') || product.id.includes('Mesoscale 2')) {
      return 'mesoscale_2'
    } else if (product.id.includes('conus') || product.id.includes('CONUS')) {
      return 'conus'
    } else if (product.region) {
      return product.region
    }
    return product.satellite || 'default'
  }
  
  // Get products in same order as displayed in sidebar
  const getOrderedProducts = (products: WeatherProduct[]): WeatherProduct[] => {
    // Group products by category/satellite (matching ProductSelector logic)
    const groups: Record<string, WeatherProduct[]> = {}
    
    products.forEach(product => {
      const category = product.satellite || product.category || 'Other'
      if (!groups[category]) {
        groups[category] = []
      }
      groups[category].push(product)
    })
    
    // Sort products within each group and flatten
    const orderedProducts: WeatherProduct[] = []
    Object.keys(groups).sort().forEach(groupName => {
      const sortedGroup = groups[groupName].sort((a, b) => a.name.localeCompare(b.name))
      orderedProducts.push(...sortedGroup)
    })
    
    return orderedProducts
  }

  // Handle product selection
  const handleProductSelect = useCallback(async (product: WeatherProduct) => {
    console.log('Selected product:', product)
    
    // Save current zoom/pan settings for the previous product group
    if (previousProductRef.current) {
      const prevGroupKey = getProductGroupKey(previousProductRef.current)
      zoomPanSettingsRef.current[prevGroupKey] = {
        zoom: zoom,
        pan: { ...panOffset }
      }
    }
    
    // Check if new product is in same group to preserve settings
    const newGroupKey = getProductGroupKey(product)
    const prevGroupKey = getProductGroupKey(previousProductRef.current)
    
    if (newGroupKey !== prevGroupKey) {
      // Different group - restore saved settings or reset
      const savedSettings = zoomPanSettingsRef.current[newGroupKey]
      if (savedSettings) {
        setZoom(savedSettings.zoom)
        setPanOffset(savedSettings.pan)
      } else {
        setZoom(1)
        setPanOffset({ x: 0, y: 0 })
      }
    }
    // Same group - keep current zoom/pan
    
    previousProductRef.current = product
    setSelectedProduct(product)
    setIsImageLoading(true)
    
    // Close sidebar on mobile after selection
    if (window.innerWidth < 1024) {
      setSidebarOpen(false)
    }
    
    try {
      const productImages = await fetchProductImages(product.id, currentDate)
      const imageArray = productImages.imageUrls || productImages.images || []
      console.log('Product images response:', productImages)
      
      if (imageArray && imageArray.length > 0) {
        // Convert image objects to URLs if needed
        const urls = imageArray.map((img: any) => {
          if (typeof img === 'string') return img
          if (img.url) return img.url
          return img
        })
        // Reverse for chronological order (oldest to newest) for animations
        const chronologicalUrls = [...urls].reverse()
        setImages(chronologicalUrls)
        setCurrentIndex(0)
        setCurrentImage(chronologicalUrls[0])
        setCurrentTimestamp(parseTimestamp(chronologicalUrls[0]))
        
        // For comparison mode, set first two images
        if (viewMode === 'compare' && chronologicalUrls.length > 1) {
          setCompareImages([chronologicalUrls[0], chronologicalUrls[1]])
          setCompareIndices([0, 1])
        }
      } else {
        // No images available for this product
        console.log('No images available for product:', product.id)
        setImages([])
        setCurrentImage(null)
        setCurrentTimestamp('')
      }
    } catch (error) {
      console.error('Error fetching product images:', error)
      setImages([])
      setCurrentImage(null)
      setCurrentTimestamp('')
    } finally {
      setIsImageLoading(false)
    }
  }, [currentDate, viewMode, zoom, panOffset])

  // Handle date change
  const handleDateChange = async (date: Date | undefined) => {
    if (!date) return
    
    setIsImageLoading(true)
    setCurrentDate(date)
    
    try {
      if (selectedProduct) {
        const productImages = await fetchProductImages(selectedProduct.id, date)
        const imageArray = productImages.imageUrls || productImages.images || []
        if (imageArray && imageArray.length > 0) {
          // Convert image objects to URLs if needed
          const urls = imageArray.map((img: any) => {
            if (typeof img === 'string') return img
            if (img.url) return img.url
            return img
          })
          // Reverse for chronological order (oldest to newest) for animations
          const chronologicalUrls = [...urls].reverse()
          setImages(chronologicalUrls)
          setCurrentIndex(0)
          setCurrentImage(chronologicalUrls[0])
          setCurrentTimestamp(parseTimestamp(chronologicalUrls[0]))
        }
      } else {
        const imageUrls = await fetchImagesByDate(date)
        const imageArray = imageUrls.imageUrls || imageUrls.images || []
        // Convert image objects to URLs if needed
        const urls = imageArray.map((img: any) => {
          if (typeof img === 'string') return img
          if (img.url) return img.url
          return img
        })
        // Reverse for chronological order (oldest to newest) for animations
        const chronologicalUrls = [...urls].reverse()
        setImages(chronologicalUrls)
        setCurrentIndex(0)
        setCurrentImage(chronologicalUrls[0])
        setCurrentTimestamp(parseTimestamp(chronologicalUrls[0]))
      }
    } catch (error) {
      console.error('Error fetching images for date:', error)
    }
  }

  // Navigation handlers
  const handlePrevious = () => {
    if (currentIndex > 0) {
      setIsImageLoading(true)
      const newIndex = currentIndex - 1
      setCurrentIndex(newIndex)
      setCurrentImage(images[newIndex])
      setCurrentTimestamp(parseTimestamp(images[newIndex]))
      // Zoom and pan are preserved to compare the same area across frames
    }
  }

  const handleNext = () => {
    if (currentIndex < images.length - 1) {
      setIsImageLoading(true)
      const newIndex = currentIndex + 1
      setCurrentIndex(newIndex)
      setCurrentImage(images[newIndex])
      setCurrentTimestamp(parseTimestamp(images[newIndex]))
      // Zoom and pan are preserved to compare the same area across frames
    }
  }

  // Animation handlers
  const handlePlayPause = () => {
    setIsAnimating(!isAnimating)
  }

  useEffect(() => {
    if (isAnimating && viewMode === 'animation') {
      animationRef.current = setInterval(() => {
        setCurrentIndex((prev) => {
          const next = prev + 1
          if (next >= images.length) {
            setIsAnimating(false)
            return 0
          }
          setCurrentImage(images[next])
          setCurrentTimestamp(parseTimestamp(images[next]))
          return next
        })
      }, animationSpeed)
    } else {
      if (animationRef.current) {
        clearInterval(animationRef.current)
      }
    }

    return () => {
      if (animationRef.current) {
        clearInterval(animationRef.current)
      }
    }
  }, [isAnimating, animationSpeed, images, viewMode])

  const handleFrameChange = (frame: number) => {
    setCurrentIndex(frame)
    setCurrentImage(images[frame])
    setCurrentTimestamp(parseTimestamp(images[frame]))
  }

  const handleReset = () => {
    setCurrentIndex(0)
    setCurrentImage(images[0])
    setCurrentTimestamp(parseTimestamp(images[0]))
    setIsAnimating(false)
    // Also reset zoom and pan when resetting to first frame
    setZoom(1)
    setPanOffset({ x: 0, y: 0 })
  }

  // Keyboard shortcuts - combined handler to avoid conflicts
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      // WASD for panning when zoomed
      if (zoom > 1 && viewMode === 'single') {
        const panSpeed = 50
        switch(e.key.toLowerCase()) {
          case 'w':
            setPanOffset(prev => ({ ...prev, y: prev.y + panSpeed }))
            e.preventDefault()
            return
          case 's':
            setPanOffset(prev => ({ ...prev, y: prev.y - panSpeed }))
            e.preventDefault()
            return
          case 'a':
            setPanOffset(prev => ({ ...prev, x: prev.x + panSpeed }))
            e.preventDefault()
            return
          case 'd':
            setPanOffset(prev => ({ ...prev, x: prev.x - panSpeed }))
            e.preventDefault()
            return
        }
      }
      
      // Other keyboard shortcuts
      switch(e.key) {
        case 'ArrowLeft':
          handlePrevious()
          break
        case 'ArrowRight':
          handleNext()
          break
        case 'ArrowUp':
          // Navigate to previous product in visual order
          if (selectedProduct && weatherProducts.length > 0) {
            // Get products in same order as sidebar (grouped and sorted)
            const orderedProducts = getOrderedProducts(weatherProducts)
            const currentIdx = orderedProducts.findIndex(p => p.id === selectedProduct.id)
            if (currentIdx > 0) {
              handleProductSelect(orderedProducts[currentIdx - 1])
            }
          }
          e.preventDefault()
          break
        case 'ArrowDown':
          // Navigate to next product in visual order
          if (selectedProduct && weatherProducts.length > 0) {
            // Get products in same order as sidebar (grouped and sorted)
            const orderedProducts = getOrderedProducts(weatherProducts)
            const currentIdx = orderedProducts.findIndex(p => p.id === selectedProduct.id)
            if (currentIdx < orderedProducts.length - 1) {
              handleProductSelect(orderedProducts[currentIdx + 1])
            }
          }
          e.preventDefault()
          break
        case ' ':
          if (viewMode === 'animation') {
            e.preventDefault()
            handlePlayPause()
          }
          break
        case '1':
          setViewMode('single')
          break
        case '2':
          setViewMode('grid')
          break
        case '3':
          setViewMode('compare')
          break
        case '4':
          setViewMode('animation')
          break
        case 'b':
          setSidebarOpen(prev => !prev)
          break
        case '+':
        case '=':
          // + key (with or without shift)
          if (viewMode === 'single') {
            setZoom(prev => Math.min(3, prev + 0.25))
          }
          break
        case '-':
          // - key for zoom out
          if (viewMode === 'single') {
            setZoom(prev => Math.max(0.5, prev - 0.25))
          }
          break
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [viewMode, zoom, handlePrevious, handleNext, handlePlayPause, selectedProduct, weatherProducts, handleProductSelect])

  // Pan and zoom handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    if (zoom > 1) {
      setIsDragging(true)
      setDragStart({ x: e.clientX - panOffset.x, y: e.clientY - panOffset.y })
      e.preventDefault()
    }
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (isDragging && zoom > 1) {
      setPanOffset({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y
      })
    }
  }

  const handleMouseUp = () => {
    setIsDragging(false)
  }

  const handleMouseLeave = () => {
    setIsDragging(false)
  }

  // Reset pan when zoom changes to 1
  useEffect(() => {
    if (zoom === 1) {
      setPanOffset({ x: 0, y: 0 })
    }
  }, [zoom])
  
  // Reset zoom/pan when changing dates (but not products - handled in handleProductSelect)
  useEffect(() => {
    setPanOffset({ x: 0, y: 0 })
    setZoom(1)
  }, [currentDate])

  const renderSingleView = () => (
    <div className="flex-1 flex flex-col items-center overflow-hidden">
      {currentImage ? (
        <div className="w-full flex-1 flex flex-col p-2 min-h-0">
          <div 
            ref={imageContainerRef}
            className="relative flex-1 min-h-0 flex items-center justify-center overflow-hidden"
            onMouseDown={handleMouseDown}
            onMouseMove={handleMouseMove}
            onMouseUp={handleMouseUp}
            onMouseLeave={handleMouseLeave}
            style={{ cursor: zoom > 1 ? (isDragging ? 'grabbing' : 'grab') : 'default' }}
          >
            {isImageLoading && (
              <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
                <div className="text-center">Loading image...</div>
              </div>
            )}
            {zoom > 1 && (
              <div className="absolute top-2 left-2 bg-black/50 text-white text-xs px-2 py-1 rounded z-10">
                Drag or use WASD keys to pan
              </div>
            )}
            <img
              key={currentImage}
              src={currentImage}
              alt="Satellite image"
              className="max-w-full max-h-full object-contain"
              style={{ 
                transform: `translate(${panOffset.x}px, ${panOffset.y}px) scale(${zoom})`,
                transition: isDragging ? 'none' : 'transform 0.2s',
                userSelect: 'none',
                pointerEvents: 'none'
              }}
              onLoad={() => setIsImageLoading(false)}
              draggable={false}
            />
          </div>
          <div className="text-center mt-2 text-sm text-muted-foreground">
            {currentTimestamp}
          </div>
        </div>
      ) : (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center text-muted-foreground">
            <p className="text-lg mb-2">No images available</p>
            <p className="text-sm">{
              selectedProduct 
                ? `No data for ${selectedProduct.name} on ${format(currentDate, 'MMM dd, yyyy')}`
                : 'Select a product from the sidebar to view images'
            }</p>
          </div>
        </div>
      )}
      
      {currentImage && (
        <div className="flex justify-center items-center gap-4 mb-2 px-2">
          {/* Navigation buttons */}
          <div className="flex gap-2">
            <Button 
              size="icon"
              variant="outline"
              onClick={handlePrevious} 
              disabled={currentIndex === 0}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button 
              size="icon"
              variant="outline"
              onClick={handleNext} 
              disabled={currentIndex === images.length - 1}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
          
          {/* Zoom controls */}
          <div className="flex gap-2">
            <Button
              size="icon"
              variant="outline"
              onClick={() => setZoom(Math.max(0.5, zoom - 0.25))}
              disabled={zoom <= 0.5}
            >
              <ZoomOut className="h-4 w-4" />
            </Button>
            <Button
              size="icon"
              variant="outline"
              onClick={() => {
                setZoom(1)
                setPanOffset({ x: 0, y: 0 })
              }}
            >
              <RefreshCw className="h-4 w-4" />
            </Button>
            <Button
              size="icon"
              variant="outline"
              onClick={() => setZoom(Math.min(3, zoom + 0.25))}
              disabled={zoom >= 3}
            >
              <ZoomIn className="h-4 w-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )

  const renderGridView = () => (
    <div className="flex-1 overflow-auto p-4">
      {images.length > 0 ? (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {images.map((image, idx) => (
          <div
            key={image}
            className={cn(
              "relative aspect-square border rounded-lg overflow-hidden cursor-pointer transition-all",
              currentIndex === idx && "ring-2 ring-primary"
            )}
            onClick={() => {
              setCurrentIndex(idx)
              setCurrentImage(image)
              setCurrentTimestamp(parseTimestamp(image))
              setViewMode('single')  // Switch to single view when clicking an image
            }}
          >
            <Image
              src={image}
              alt={`Image ${idx + 1}`}
              fill
              style={{ objectFit: 'cover' }}
              sizes="(max-width: 768px) 50vw, (max-width: 1024px) 33vw, 25vw"
            />
            <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-xs p-1 text-center">
              {parseTimestamp(image)}
            </div>
          </div>
          ))}
        </div>
      ) : (
        <div className="flex h-full items-center justify-center">
          <div className="text-center text-muted-foreground">
            <p className="text-lg mb-2">No images available</p>
            <p className="text-sm">{
              selectedProduct 
                ? `No data for ${selectedProduct.name}`
                : 'Select a product from the sidebar to view images'
            }</p>
          </div>
        </div>
      )}
    </div>
  )

  const handleCompareIndexChange = (imageIndex: number, value: number) => {
    const newIndices: [number, number] = [...compareIndices] as [number, number]
    newIndices[imageIndex] = value
    setCompareIndices(newIndices)
    
    if (images[newIndices[0]] && images[newIndices[1]]) {
      setCompareImages([images[newIndices[0]], images[newIndices[1]]])
    }
  }

  const renderCompareView = () => (
    <div className="flex-1 flex flex-col p-4 min-h-0">
      {images.length >= 2 ? (
        <>
          {/* Selection controls */}
          <div className="flex gap-4 mb-4">
            <div className="flex-1 flex items-center gap-2">
              <label className="text-sm font-medium whitespace-nowrap">Left image:</label>
              <select
                value={compareIndices[0]}
                onChange={(e) => handleCompareIndexChange(0, parseInt(e.target.value))}
                className="flex-1 px-3 py-1 text-sm border rounded-md bg-background"
              >
                {images.map((_, idx) => (
                  <option key={idx} value={idx}>
                    {parseTimestamp(images[idx])}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex-1 flex items-center gap-2">
              <label className="text-sm font-medium whitespace-nowrap">Right image:</label>
              <select
                value={compareIndices[1]}
                onChange={(e) => handleCompareIndexChange(1, parseInt(e.target.value))}
                className="flex-1 px-3 py-1 text-sm border rounded-md bg-background"
              >
                {images.map((_, idx) => (
                  <option key={idx} value={idx}>
                    {parseTimestamp(images[idx])}
                  </option>
                ))}
              </select>
            </div>
          </div>
          
          {/* Image comparison */}
          <div className="flex-1 flex gap-2 min-h-0">
            {compareImages.length >= 2 ? (
              <>
                <div className="flex-1 min-h-0 flex items-center justify-center relative border rounded-lg overflow-hidden">
                  <img
                    src={compareImages[0]}
                    alt="Compare image 1"
                    className="max-w-full max-h-full object-contain"
                  />
                  <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-xs p-1 text-center">
                    {parseTimestamp(compareImages[0])}
                  </div>
                </div>
                <div className="flex-1 min-h-0 flex items-center justify-center relative border rounded-lg overflow-hidden">
                  <img
                    src={compareImages[1]}
                    alt="Compare image 2"
                    className="max-w-full max-h-full object-contain"
                  />
                  <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-xs p-1 text-center">
                    {parseTimestamp(compareImages[1])}
                  </div>
                </div>
              </>
            ) : (
              <div className="flex-1 flex items-center justify-center text-muted-foreground">
                Loading comparison images...
              </div>
            )}
          </div>
        </>
      ) : (
        <div className="flex-1 flex items-center justify-center text-muted-foreground">
          <div className="text-center">
            <p className="text-lg mb-2">Not enough images to compare</p>
            <p className="text-sm">At least 2 images are required for comparison mode</p>
          </div>
        </div>
      )}
    </div>
  )

  const renderAnimationView = () => (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="flex-1 min-h-0 flex items-center justify-center p-4 overflow-hidden">
        {currentImage && (
          <div className="relative flex-1 min-h-0 max-h-full flex items-center justify-center">
            <img
              key={currentImage}
              src={currentImage}
              alt="Animation frame"
              className="max-w-full max-h-full object-contain"
            />
            <div className="absolute bottom-4 left-4 bg-black/50 text-white text-sm px-2 py-1 rounded">
              {currentTimestamp}
            </div>
          </div>
        )}
      </div>
      <AnimationControls
        isPlaying={isAnimating}
        onPlayPause={handlePlayPause}
        onStepBackward={handlePrevious}
        onStepForward={handleNext}
        onReset={handleReset}
        currentFrame={currentIndex}
        totalFrames={images.length}
        onFrameChange={handleFrameChange}
        speed={animationSpeed}
        onSpeedChange={setAnimationSpeed}
        canStepBackward={currentIndex > 0}
        canStepForward={currentIndex < images.length - 1}
      />
    </div>
  )


  return (
    <div className="h-full w-full flex flex-col">
      {/* Header */}
      <div className="border-b px-4 py-2 flex items-center justify-between relative z-30">
        <div className="flex items-center gap-2">
          <Button
            size="icon"
            variant="ghost"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className=""
          >
            {sidebarOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </Button>
          
          <h1 className="text-lg font-semibold hidden sm:block">GOES Data Viewer</h1>
        </div>
        
        <div className="flex items-center gap-1 sm:gap-2">
          <div className="hidden sm:block">
            <ViewModeSelector viewMode={viewMode} onViewModeChange={setViewMode} />
          </div>
          
          {/* Mobile view mode selector */}
          <Popover>
            <PopoverTrigger asChild className="sm:hidden">
              <Button variant="outline" size="icon">
                <Maximize2 className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-48 p-2" align="end">
              <div className="space-y-1">
                <Button variant={viewMode === 'single' ? 'secondary' : 'ghost'} size="sm" className="w-full justify-start" onClick={() => setViewMode('single')}>
                  <Maximize2 className="h-4 w-4 mr-2" /> Single View
                </Button>
                <Button variant={viewMode === 'grid' ? 'secondary' : 'ghost'} size="sm" className="w-full justify-start" onClick={() => setViewMode('grid')}>
                  <Grid3X3 className="h-4 w-4 mr-2" /> Grid View
                </Button>
                <Button variant={viewMode === 'compare' ? 'secondary' : 'ghost'} size="sm" className="w-full justify-start" onClick={() => setViewMode('compare')}>
                  <SplitSquareHorizontal className="h-4 w-4 mr-2" /> Compare
                </Button>
                <Button variant={viewMode === 'animation' ? 'secondary' : 'ghost'} size="sm" className="w-full justify-start" onClick={() => setViewMode('animation')}>
                  <Play className="h-4 w-4 mr-2" /> Animation
                </Button>
              </div>
            </PopoverContent>
          </Popover>
          
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="icon">
                <CalendarIcon className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <Calendar
                mode="single"
                selected={currentDate}
                onSelect={handleDateChange}
                disabled={(date) => 
                  !availableDates.some(availableDate => 
                    availableDate.toDateString() === date.toDateString()
                  )
                }
              />
            </PopoverContent>
          </Popover>
          
          <Link href="/emwin">
            <Button variant="outline" size="icon" title="EMWIN Text Data">
              <FileText className="h-4 w-4" />
            </Button>
          </Link>
          
          <ThemeToggle />
          
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="icon">
                <FileQuestion className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-4" align="end">
              <div className="space-y-2 text-sm">
                <p className="font-semibold">GOES Data Viewer</p>
                <p>A modern interface for viewing GOES satellite imagery and weather data.</p>
                <p>Features:</p>
                <ul className="list-disc list-inside ml-2">
                  <li>Multiple view modes (single, grid, compare, animation)</li>
                  <li>Support for all GOES channels and products</li>
                  <li>EMWIN text data viewer</li>
                  <li>Himawari satellite support</li>
                </ul>
                <p className="text-xs text-muted-foreground mt-2">
                  Hotkeys: B (sidebar), 1-4 (views), ←→ (navigate), +/- (zoom), WASD (pan)
                </p>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
      
      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <div className={cn(
          "bg-background border-r overflow-hidden shadow-xl transition-all duration-300",
          "fixed lg:relative top-[57px] lg:top-0 bottom-0 left-0",
          "z-[100]",
          sidebarOpen ? "w-80" : "w-0"
        )}>
          {sidebarOpen && (
            <ProductSelector
              products={weatherProducts}
              selectedProduct={selectedProduct}
              onProductSelect={handleProductSelect}
            />
          )}
        </div>
        
        {/* Content area */}
        <div className={cn(
          "flex-1 flex flex-col relative",
          sidebarOpen && "lg:flex hidden"
        )}>
          {viewMode === 'single' && renderSingleView()}
          {viewMode === 'grid' && renderGridView()}
          {viewMode === 'compare' && renderCompareView()}
          {viewMode === 'animation' && renderAnimationView()}
        </div>
      </div>
    </div>
  )
}