'use client'

import { useState } from 'react'
import { EmwinViewer } from '@/components/emwin-viewer'
import { Button } from '@/components/ui/button'
import { ThemeToggle } from '@/components/ui/theme-toggle'
import { ArrowLeft, CalendarIcon } from 'lucide-react'
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
          <Link href="/">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <h1 className="text-lg font-semibold">EMWIN Text Data Viewer</h1>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="text-sm text-muted-foreground mr-2">
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
          
          <ThemeToggle />
        </div>
      </div>
      
      {/* EMWIN Viewer */}
      <div className="flex-1 overflow-hidden">
        <EmwinViewer selectedDate={selectedDate} />
      </div>
    </div>
  )
}