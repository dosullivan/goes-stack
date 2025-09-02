'use client'

import { useState } from 'react'
import { EmwinViewer } from '@/components/emwin-viewer'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { CalendarIcon, Menu, X, Globe, FileQuestion } from 'lucide-react'
import Link from 'next/link'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Calendar } from '@/components/ui/calendar'
import { format } from 'date-fns'

export default function EmwinPage() {
  const [selectedDate, setSelectedDate] = useState<Date>(new Date())
  const [sidebarOpen, setSidebarOpen] = useState(true)

  const handleDateChange = (date: Date | undefined) => {
    if (date) {
      setSelectedDate(date)
    }
  }

  return (
    <div className="h-screen w-full flex flex-col">
      {/* Header */}
      <div className="border-b px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Button
            size="icon"
            variant="ghost"
            onClick={() => setSidebarOpen(!sidebarOpen)}
          >
            {sidebarOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </Button>
          
          <h1 className="text-lg font-semibold hidden sm:block">EMWIN Text Data Viewer</h1>
        </div>
        
        <div className="flex items-center gap-1 sm:gap-2">
          <div className="text-sm text-muted-foreground mr-2 hidden sm:block">
            {format(selectedDate, 'MMM dd, yyyy')}
          </div>
          
          <Popover>
            <PopoverTrigger asChild>
              <Button variant="outline" size="icon">
                <CalendarIcon className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="end">
              <Calendar
                mode="single"
                selected={selectedDate}
                onSelect={handleDateChange}
                disabled={(date) => date > new Date() || date < new Date('2024-01-01')}
              />
            </PopoverContent>
          </Popover>
          
          <Link href="/">
            <Button variant="outline" size="icon" title="Back to GOES Viewer">
              <Globe className="h-4 w-4" />
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
                <p className="font-semibold">EMWIN Text Data Viewer</p>
                <p>View and browse EMWIN (Emergency Managers Weather Information Network) text products.</p>
                <p>Features:</p>
                <ul className="list-disc list-inside ml-2">
                  <li>Weather forecasts and warnings</li>
                  <li>Station observations</li>
                  <li>Climate summaries</li>
                  <li>Filter by office location</li>
                </ul>
                <p className="text-xs text-muted-foreground mt-2">
                  Hotkeys: B (toggle sidebar), W/S (previous/next file)
                </p>
                <p className="text-xs text-muted-foreground">
                  Data updates every few minutes
                </p>
              </div>
            </PopoverContent>
          </Popover>
        </div>
      </div>
      
      {/* EMWIN Viewer */}
      <div className="flex-1 overflow-hidden">
        <EmwinViewer selectedDate={selectedDate} sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />
      </div>
    </div>
  )
}