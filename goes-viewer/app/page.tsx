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
  RefreshCw
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
import { EmwinViewer } from '@/components/emwin-viewer'

type ViewMode = 'single' | 'grid' | 'compare' | 'animation' | 'emwin'

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
  const [zoom, setZoom] = useState(1)
  const animationRef = useRef<NodeJS.Timeout | null>(null)

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
        
        // Set default product - try radar or other EMWIN products first since FD might not have data
        const defaultProduct = parsedProducts.find(p => 
          p.id.includes('radar_us') || 
          p.id.includes('radar_') ||
          p.category === 'emwin'
        ) || parsedProducts.find(p => 
          p.id === 'fd_color' || 
          p.id === 'fc' || 
          p.name.toLowerCase().includes('color')
        ) || parsedProducts[0]
        if (defaultProduct) {
          setSelectedProduct(defaultProduct)
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
                setImages(urls)
                setCurrentIndex(0)
                setCurrentImage(urls[0])
                setCurrentTimestamp(parseTimestamp(urls[0]))
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

  // Handle product selection
  const handleProductSelect = useCallback(async (product: WeatherProduct) => {
    console.log('Selected product:', product)
    setSelectedProduct(product)
    setIsImageLoading(true)
    
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
        setImages(urls)
        setCurrentIndex(0)
        setCurrentImage(urls[0])
        setCurrentTimestamp(parseTimestamp(urls[0]))
        
        // For comparison mode, set first two images
        if (viewMode === 'compare' && urls.length > 1) {
          setCompareImages([urls[0], urls[1]])
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
  }, [currentDate, viewMode])

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
          setImages(urls)
          setCurrentIndex(0)
          setCurrentImage(urls[0])
          setCurrentTimestamp(parseTimestamp(urls[0]))
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
        setImages(urls)
        setCurrentIndex(0)
        setCurrentImage(urls[0])
        setCurrentTimestamp(parseTimestamp(urls[0]))
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
    }
  }

  const handleNext = () => {
    if (currentIndex < images.length - 1) {
      setIsImageLoading(true)
      const newIndex = currentIndex + 1
      setCurrentIndex(newIndex)
      setCurrentImage(images[newIndex])
      setCurrentTimestamp(parseTimestamp(images[newIndex]))
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
  }

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      switch(e.key) {
        case 'ArrowLeft':
          handlePrevious()
          break
        case 'ArrowRight':
          handleNext()
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
        case '5':
          setViewMode('emwin')
          break
        case 's':
          setSidebarOpen(!sidebarOpen)
          break
      }
    }

    window.addEventListener('keydown', handleKeyPress)
    return () => window.removeEventListener('keydown', handleKeyPress)
  }, [currentIndex, images, viewMode, sidebarOpen])

  const renderSingleView = () => (
    <div className="flex-1 flex flex-col items-center">
      {currentImage ? (
        <div className="relative w-full max-w-6xl flex-1 mb-2 flex flex-col">
          <div className="relative flex-1">
            {isImageLoading && (
              <div className="absolute inset-0 flex items-center justify-center bg-background/50 z-10">
                <div className="text-center">Loading image...</div>
              </div>
            )}
            <Image
              key={currentImage}
              src={currentImage}
              alt="Satellite image"
              fill
              priority
              style={{ 
                objectFit: 'contain',
                transform: `scale(${zoom})`
              }}
              onLoadingComplete={() => setIsImageLoading(false)}
              onLoad={() => setIsImageLoading(false)}
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
        <div className="flex justify-center items-center gap-4 mb-2">
        <Button 
          className="w-32" 
          onClick={handlePrevious} 
          disabled={currentIndex === 0}
        >
          <ChevronLeft className="mr-2 h-4 w-4" /> Previous
        </Button>
        
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
            onClick={() => setZoom(1)}
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
        
        <Button 
          className="w-32" 
          onClick={handleNext} 
          disabled={currentIndex === images.length - 1}
        >
          Next <ChevronRight className="ml-2 h-4 w-4" />
        </Button>
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

  const renderCompareView = () => (
    <div className="flex-1 flex gap-2 p-4">
      {compareImages.length >= 2 ? (
        <>
          <div className="flex-1 relative border rounded-lg overflow-hidden">
            <Image
              src={compareImages[0]}
              alt="Compare image 1"
              fill
              style={{ objectFit: 'contain' }}
            />
            <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-xs p-1 text-center">
              {parseTimestamp(compareImages[0])}
            </div>
          </div>
          <div className="flex-1 relative border rounded-lg overflow-hidden">
            <Image
              src={compareImages[1]}
              alt="Compare image 2"
              fill
              style={{ objectFit: 'contain' }}
            />
            <div className="absolute bottom-0 left-0 right-0 bg-black/50 text-white text-xs p-1 text-center">
              {parseTimestamp(compareImages[1])}
            </div>
          </div>
        </>
      ) : (
        <div className="flex-1 flex items-center justify-center text-muted-foreground">
          Select images to compare
        </div>
      )}
    </div>
  )

  const renderAnimationView = () => (
    <div className="flex-1 flex flex-col">
      <div className="flex-1 flex items-center justify-center p-4">
        {currentImage && (
          <div className="relative w-full max-w-4xl aspect-square">
            <Image
              key={currentImage}
              src={currentImage}
              alt="Animation frame"
              fill
              style={{ objectFit: 'contain' }}
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

  const renderEmwinView = () => (
    <div className="flex-1 overflow-hidden">
      <EmwinViewer />
    </div>
  )

  return (
    <div className="h-screen w-full flex flex-col">
      {/* Header */}
      <div className="border-b px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            size="icon"
            variant="ghost"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            className="lg:hidden"
          >
            {sidebarOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </Button>
          
          <h1 className="text-lg font-semibold">GOES Stack Viewer</h1>
        </div>
        
        <div className="flex items-center gap-2">
          <ViewModeSelector viewMode={viewMode} onViewModeChange={setViewMode} />
          
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
          
          <ThemeToggle />
          
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="icon">
                <FileQuestion className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-4" align="end">
              <div className="space-y-2 text-sm">
                <p className="font-semibold">GOES Stack Viewer</p>
                <p>A modern interface for viewing GOES satellite imagery and weather data.</p>
                <p>Features:</p>
                <ul className="list-disc list-inside ml-2">
                  <li>Multiple view modes (single, grid, compare, animation, EMWIN)</li>
                  <li>Support for all GOES channels and products</li>
                  <li>EMWIN text data viewer</li>
                  <li>Himawari satellite support</li>
                </ul>
                <p className="text-xs text-muted-foreground mt-2">
                  Press 'S' to toggle sidebar, 1-5 for view modes, arrow keys to navigate
                </p>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
      
      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar - only show for image modes */}
        {viewMode !== 'emwin' && (
          <div className={cn(
            "border-r bg-muted/10 transition-all duration-300",
            sidebarOpen ? "w-64" : "w-0",
            "lg:w-64"
          )}>
            {sidebarOpen && (
              <ProductSelector
                products={weatherProducts}
                selectedProduct={selectedProduct}
                onProductSelect={handleProductSelect}
              />
            )}
          </div>
        )}
        
        {/* Content area */}
        <div className="flex-1 flex flex-col">
          {viewMode === 'single' && renderSingleView()}
          {viewMode === 'grid' && renderGridView()}
          {viewMode === 'compare' && renderCompareView()}
          {viewMode === 'animation' && renderAnimationView()}
          {viewMode === 'emwin' && renderEmwinView()}
        </div>
      </div>
    </div>
  )
}