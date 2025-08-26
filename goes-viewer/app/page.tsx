'use client'

import { useState, useEffect } from 'react'
import Image from 'next/image'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from '@/components/ui/calendar'
import { ChevronLeft, ChevronRight, CalendarIcon, FileQuestion } from 'lucide-react'
import { fetchLatestImage, fetchAvailableDates, fetchImagesByDate } from '@/lib/api'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { parseTimestamp } from '@/lib/utils'

export default function Home() {
  const [currentImage, setCurrentImage] = useState<string | null>(null)
  const [currentDate, setCurrentDate] = useState<Date>(new Date())
  const [availableDates, setAvailableDates] = useState<Date[]>([])
  const [images, setImages] = useState<string[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [isImageLoading, setIsImageLoading] = useState(true)
  const [currentTimestamp, setCurrentTimestamp] = useState<string>('')

  useEffect(() => {
    const fetchInitialData = async () => {
      try {
        const latestImage = await fetchLatestImage()
        const dates = await fetchAvailableDates()
        setAvailableDates(dates.availableDates)
        
        if (dates.availableDates.length > 0) {
          const latestDate = dates.availableDates[0]
          setCurrentDate(latestDate)
          const imageUrls = await fetchImagesByDate(latestDate)
          setImages(imageUrls.imageUrls)
          
          const latestImageIndex = imageUrls.imageUrls.indexOf(latestImage.imageUrl)
          setCurrentIndex(latestImageIndex)
          setCurrentImage(latestImage.imageUrl)
        }
      } catch (error) {
        console.error('Error fetching initial data:', error)
      }
    }
    fetchInitialData()
  }, [])

  const handleDateChange = async (date: Date | undefined) => {
    if (!date) return;
    
    setIsImageLoading(true)
    try {
      setCurrentDate(date)
      const imageUrls = await fetchImagesByDate(date)
      setImages(imageUrls.imageUrls)
      setCurrentIndex(0)
      setCurrentImage(imageUrls.imageUrls[0])
      setCurrentTimestamp(parseTimestamp(imageUrls.imageUrls[0]))
    } catch (error) {
      console.error('Error fetching images for date:', error)
    }
  }

  const handlePrevious = () => {
    if (currentIndex > 0) {
      setIsImageLoading(true)
      setCurrentIndex(currentIndex - 1)
      const prevImage = images[currentIndex - 1]
      setCurrentImage(prevImage)
      setCurrentTimestamp(parseTimestamp(prevImage))
    }
  }

  const handleNext = () => {
    if (currentIndex < images.length - 1) {
      setIsImageLoading(true)
      setCurrentIndex(currentIndex + 1)
      const nextImage = images[currentIndex + 1]
      setCurrentImage(nextImage)
      setCurrentTimestamp(parseTimestamp(nextImage))
    }
  }

  return (
    <div className="h-[calc(100%-3rem)] w-full flex flex-col">
      <div className="mb-4 flex items-center space-x-2">
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline" size="icon">
              <CalendarIcon className="h-4 w-4" />
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0 bg-popover border shadow-md z-50" align="start">
            <Calendar
              mode="single"
              selected={currentDate}
              onSelect={handleDateChange}
              disabled={(date) => 
                !availableDates.some(availableDate => 
                  availableDate.toDateString() === date.toDateString()
                )
              }
              className="rounded-md"
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
          <PopoverContent 
            className="w-[calc(100vw-2rem)] sm:w-auto max-w-[20rem] p-4 bg-popover border shadow-md z-50" 
            align="start"
            side="bottom"
          >
            <div className="space-y-2 text-sm">
              <p>This site serves <a href="https://www.goes-r.gov/spacesegment/spacecraft.html" target="_blank" rel="noopener noreferrer" className="underline">GOES-19</a> images captured by a Raspberry Pi.</p>
              <p>The images are pulled from the NOAA GOES-19 satellite and are updated every 30 minutes, during daylight hours.</p>
              <p>Because the Pi is solar powered, images are only available weather permitting, and may not be available for every day.</p>
            </div>
          </PopoverContent>
        </Popover>
      </div>

      <div className="flex-1 flex flex-col items-center">
        {currentImage && (
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
                alt="Current image"
                fill
                priority
                style={{ objectFit: 'contain' }}
                onLoadingComplete={() => setIsImageLoading(false)}
                onLoad={() => setIsImageLoading(false)}
              />
            </div>
            <div className="text-center mt-2">{currentTimestamp}</div>
          </div>
        )}

        <div className="flex justify-center space-x-4 mb-2">
          <Button className="w-32" onClick={handlePrevious} disabled={currentIndex === 0}>
            <ChevronLeft className="mr-2 h-4 w-4" /> Previous
          </Button>
          <Button className="w-32" onClick={handleNext} disabled={currentIndex === images.length - 1}>
            Next <ChevronRight className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}

