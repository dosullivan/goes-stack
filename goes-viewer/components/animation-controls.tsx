'use client'

import { Button } from '@/components/ui/button'
import { Slider } from '@/components/ui/slider'
import { 
  Play, 
  Pause, 
  SkipBack, 
  SkipForward,
  RotateCcw
} from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface AnimationControlsProps {
  isPlaying: boolean
  onPlayPause: () => void
  onStepBackward: () => void
  onStepForward: () => void
  onReset: () => void
  currentFrame: number
  totalFrames: number
  onFrameChange: (frame: number) => void
  speed: number
  onSpeedChange: (speed: number) => void
  canStepBackward: boolean
  canStepForward: boolean
}

export function AnimationControls({
  isPlaying,
  onPlayPause,
  onStepBackward,
  onStepForward,
  onReset,
  currentFrame,
  totalFrames,
  onFrameChange,
  speed,
  onSpeedChange,
  canStepBackward,
  canStepForward
}: AnimationControlsProps) {
  const speedOptions = [
    { value: 100, label: '10 fps' },
    { value: 200, label: '5 fps' },
    { value: 500, label: '2 fps' },
    { value: 1000, label: '1 fps' },
    { value: 2000, label: '0.5 fps' }
  ]

  return (
    <div className="flex flex-col gap-4 p-4 border-t bg-background">
      <div className="flex items-center justify-center gap-2">
        <Button
          size="icon"
          variant="outline"
          onClick={onReset}
          disabled={currentFrame === 0}
        >
          <RotateCcw className="h-4 w-4" />
        </Button>
        
        <Button
          size="icon"
          variant="outline"
          onClick={onStepBackward}
          disabled={!canStepBackward}
        >
          <SkipBack className="h-4 w-4" />
        </Button>
        
        <Button
          size="icon"
          onClick={onPlayPause}
        >
          {isPlaying ? (
            <Pause className="h-4 w-4" />
          ) : (
            <Play className="h-4 w-4" />
          )}
        </Button>
        
        <Button
          size="icon"
          variant="outline"
          onClick={onStepForward}
          disabled={!canStepForward}
        >
          <SkipForward className="h-4 w-4" />
        </Button>

        <Select value={speed.toString()} onValueChange={(value) => onSpeedChange(Number(value))}>
          <SelectTrigger className="w-24">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {speedOptions.map(option => (
              <SelectItem key={option.value} value={option.value.toString()}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      
      <div className="flex items-center gap-4">
        <span className="text-sm text-muted-foreground min-w-12">
          {currentFrame + 1}/{totalFrames}
        </span>
        <Slider
          value={[currentFrame]}
          onValueChange={(value) => onFrameChange(value[0])}
          max={totalFrames - 1}
          min={0}
          step={1}
          className="flex-1"
        />
      </div>
    </div>
  )
}