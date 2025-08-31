'use client'

import { useState, useEffect } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { fetchEmwinCategories, fetchEmwinFiles, fetchEmwinContent, fetchWeatherOffices } from '@/lib/api'
import { format } from 'date-fns'
import { FileText, Calendar, MapPin, RefreshCw } from 'lucide-react'

interface EmwinFile {
  key?: string
  url?: string
  filename: string
  station?: string
  timestamp?: string
  size?: number
}

interface EmwinViewerProps {
  selectedDate?: Date
}

export function EmwinViewer({ selectedDate: propSelectedDate }: EmwinViewerProps) {
  const [categories, setCategories] = useState<any[]>([])
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [offices, setOffices] = useState<any[]>([])
  const [selectedOffice, setSelectedOffice] = useState<string>('')
  const [files, setFiles] = useState<EmwinFile[]>([])
  const [selectedFile, setSelectedFile] = useState<EmwinFile | null>(null)
  const [fileContent, setFileContent] = useState<string>('')
  const [isLoading, setIsLoading] = useState(false)
  const [selectedDate, setSelectedDate] = useState<Date>(propSelectedDate || new Date())

  useEffect(() => {
    loadCategories()
    loadOffices()
  }, [])

  useEffect(() => {
    if (propSelectedDate) {
      setSelectedDate(propSelectedDate)
    }
  }, [propSelectedDate])

  const loadCategories = async () => {
    try {
      const data = await fetchEmwinCategories()
      setCategories(data.categories || [])
      if (data.categories && data.categories.length > 0) {
        setSelectedCategory(data.categories[0].key)
      }
    } catch (error) {
      console.error('Error loading EMWIN categories:', error)
    }
  }

  const loadOffices = async () => {
    try {
      const data = await fetchWeatherOffices()
      setOffices(data.offices || [])
    } catch (error) {
      console.error('Error loading weather offices:', error)
    }
  }

  useEffect(() => {
    if (selectedCategory) {
      loadFiles()
    }
  }, [selectedCategory, selectedDate, selectedOffice])

  const loadFiles = async () => {
    setIsLoading(true)
    try {
      console.log('Loading files for category:', selectedCategory, 'date:', selectedDate, 'office:', selectedOffice)
      const data = await fetchEmwinFiles(selectedCategory, selectedDate, undefined, selectedOffice || undefined)
      console.log('EMWIN files response:', data)
      const fileList = data.files || []
      setFiles(fileList)
      setSelectedFile(null)
      setFileContent('')
      console.log(`Loaded ${fileList.length} files`)
    } catch (error) {
      console.error('Error loading EMWIN files:', error)
      setFiles([])
    } finally {
      setIsLoading(false)
    }
  }

  const loadFileContent = async (file: EmwinFile) => {
    setSelectedFile(file)
    setIsLoading(true)
    try {
      // Extract the object key from the URL if needed
      let key = file.key
      if (!key && file.url) {
        // Extract path after the base URL
        // URL format: https://example.com/bucket-name/emwin/2024-12-25/filename.TXT
        const urlParts = file.url.split('/')
        const emwinIndex = urlParts.indexOf('emwin')
        if (emwinIndex !== -1) {
          key = urlParts.slice(emwinIndex).join('/')
        } else {
          // Try to extract everything after the domain
          const match = file.url.match(/https?:\/\/[^\/]+\/(.*)/)
          if (match) {
            key = match[1]
          }
        }
      }
      
      if (!key) {
        console.error('Could not extract key from file:', file)
        throw new Error('No valid key found for file')
      }
      
      console.log('Loading content for key:', key)
      
      const data = await fetchEmwinContent(key)
      setFileContent(data.content || '')
    } catch (error) {
      console.error('Error loading file content:', error)
      setFileContent('Error loading file content')
    } finally {
      setIsLoading(false)
    }
  }

  const formatFileInfo = (file: EmwinFile) => {
    // Use the station from the API response if available
    const station = file.station || 'Unknown'
    
    // Parse the timestamp if available
    let time = 'Unknown'
    if (file.timestamp) {
      try {
        const date = new Date(file.timestamp)
        if (!isNaN(date.getTime())) {
          time = format(date, 'MMM dd, HH:mm')
        }
      } catch (e) {
        // If parsing fails, try extracting from filename
        const match = file.filename.match(/(\d{14})/)
        if (match) {
          try {
            const year = parseInt(match[1].substring(0, 4))
            const month = parseInt(match[1].substring(4, 6)) - 1
            const day = parseInt(match[1].substring(6, 8))
            const hour = parseInt(match[1].substring(8, 10))
            const minute = parseInt(match[1].substring(10, 12))
            const date = new Date(year, month, day, hour, minute)
            if (!isNaN(date.getTime())) {
              time = format(date, 'MMM dd, HH:mm')
            }
          } catch (e2) {
            // Keep as 'Unknown'
          }
        }
      }
    }
    
    return {
      station,
      time
    }
  }

  return (
    <div className="h-full flex flex-col p-4">
      <div className="mb-4 flex gap-4 items-center">
        <Select value={selectedCategory} onValueChange={setSelectedCategory}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder="Select category" />
          </SelectTrigger>
          <SelectContent>
            {categories.map(category => (
              <SelectItem key={category.key} value={category.key}>
                {category.title || category.key}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={selectedOffice || "all"} onValueChange={(value) => setSelectedOffice(value === "all" ? "" : value)}>
          <SelectTrigger className="w-64">
            <SelectValue placeholder="All offices" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All offices</SelectItem>
            {offices.map(office => (
              <SelectItem key={office.stationId} value={office.stationId}>
                {office.city}, {office.state} ({office.stationId})
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Button
          size="icon"
          variant="outline"
          onClick={loadFiles}
          disabled={isLoading}
        >
          <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      <div className="flex-1 flex gap-4 overflow-hidden">
        <Card className="w-1/3 flex flex-col">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Available Files</CardTitle>
            <CardDescription className="text-xs">
              {files.length} files in {selectedCategory}
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden p-0">
            <ScrollArea className="h-full px-4">
              <div className="space-y-2 pb-4">
                {files.map(file => {
                  const info = formatFileInfo(file)
                  return (
                    <Button
                      key={file.url || file.filename}
                      variant={selectedFile?.url === file.url ? "secondary" : "ghost"}
                      className="w-full justify-start h-auto py-2 px-3"
                      onClick={() => loadFileContent(file)}
                    >
                      <div className="flex flex-col items-start w-full">
                        <div className="flex items-center gap-2 text-sm">
                          <FileText className="h-3 w-3" />
                          <span className="truncate">{file.filename}</span>
                        </div>
                        <div className="flex items-center gap-3 text-xs text-muted-foreground mt-1">
                          <span className="flex items-center gap-1">
                            <MapPin className="h-3 w-3" />
                            {info.station}
                          </span>
                          <span className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {info.time}
                          </span>
                        </div>
                      </div>
                    </Button>
                  )
                })}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>

        <Card className="flex-1 flex flex-col">
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">
              {selectedFile ? selectedFile.filename : 'File Content'}
            </CardTitle>
            {selectedFile && (
              <CardDescription className="text-xs">
                {selectedFile.size ? `${(selectedFile.size / 1024).toFixed(1)} KB` : ''}
              </CardDescription>
            )}
          </CardHeader>
          <CardContent className="flex-1 overflow-hidden p-0">
            <ScrollArea className="h-full">
              {isLoading ? (
                <div className="p-4 text-center text-muted-foreground">
                  Loading content...
                </div>
              ) : fileContent ? (
                <pre className="p-4 text-xs font-mono whitespace-pre-wrap">
                  {fileContent}
                </pre>
              ) : (
                <div className="p-4 text-center text-muted-foreground">
                  Select a file to view its content
                </div>
              )}
            </ScrollArea>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}