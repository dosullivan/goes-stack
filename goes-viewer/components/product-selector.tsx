'use client'

import { useState, useEffect } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { ChevronRight, ChevronDown, Satellite, Cloud, Radar, FileText } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Product {
  id: string
  name: string
  category: string
  satellite?: string
  region?: string
  channel?: string
  type?: string
}

interface ProductGroup {
  name: string
  icon: React.ReactNode
  products: Product[]
  expanded: boolean
}

interface ProductSelectorProps {
  onProductSelect: (product: Product) => void
  selectedProduct: Product | null
  products: Product[]
}

export function ProductSelector({ onProductSelect, selectedProduct, products }: ProductSelectorProps) {
  const [productGroups, setProductGroups] = useState<ProductGroup[]>([])

  useEffect(() => {
    // Organize products into groups
    const groups: Record<string, Product[]> = {}
    
    if (products && Array.isArray(products)) {
      products.forEach(product => {
        const category = product.satellite || product.category || 'Other'
        if (!groups[category]) {
          groups[category] = []
        }
        groups[category].push(product)
      })
    }

    // Create group objects with icons
    const groupedProducts: ProductGroup[] = Object.entries(groups).map(([name, items]) => ({
      name,
      icon: getIconForCategory(name),
      products: items.sort((a, b) => a.name.localeCompare(b.name)),
      expanded: name === 'goes19' // Default expand GOES-19
    }))

    setProductGroups(groupedProducts)
  }, [products])

  const getIconForCategory = (category: string): React.ReactNode => {
    const lowerCategory = category.toLowerCase()
    if (lowerCategory.includes('goes') || lowerCategory.includes('himawari')) {
      return <Satellite className="h-4 w-4" />
    } else if (lowerCategory.includes('radar')) {
      return <Radar className="h-4 w-4" />
    } else if (lowerCategory.includes('emwin') || lowerCategory.includes('text')) {
      return <FileText className="h-4 w-4" />
    } else {
      return <Cloud className="h-4 w-4" />
    }
  }

  const toggleGroup = (groupName: string) => {
    setProductGroups(groups =>
      groups.map(group =>
        group.name === groupName
          ? { ...group, expanded: !group.expanded }
          : group
      )
    )
  }

  const formatProductName = (product: Product): string => {
    // If it's already a well-formatted name (from backend), use it directly
    if (product.name && !product.name.includes('Unknown')) {
      return product.name
    }
    // Otherwise, try to format based on channel/region info
    if (product.channel) {
      return `Channel ${product.channel.replace('ch', '')}${product.region ? ` - ${product.region.toUpperCase()}` : ''}`
    }
    return product.name
  }

  const getProductDescription = (product: Product): string | null => {
    // Return null to use backend descriptions instead of our manual mapping
    return null
  }

  return (
    <div className="h-full flex flex-col bg-background">
      <div className="p-4 border-b bg-background">
        <h2 className="text-lg font-semibold">Data Products</h2>
      </div>
      
      <div className="flex-1 overflow-y-auto scrollbar-gutter-stable bg-background">
        <div className="p-2">
          {productGroups.map(group => (
            <div key={group.name} className="mb-2">
              <Button
                variant="ghost"
                size="sm"
                className="w-full justify-start px-2"
                onClick={() => toggleGroup(group.name)}
              >
                <div className="flex items-center w-full">
                  {group.expanded ? (
                    <ChevronDown className="h-4 w-4 mr-2 flex-shrink-0" />
                  ) : (
                    <ChevronRight className="h-4 w-4 mr-2 flex-shrink-0" />
                  )}
                  {group.icon}
                  <span className="ml-2 font-medium capitalize flex-1 text-left truncate">
                    {group.name.replace(/([A-Z])/g, ' $1').trim()}
                  </span>
                  <span className="text-xs text-muted-foreground flex-shrink-0">
                    {group.products.length}
                  </span>
                </div>
              </Button>
              
              {group.expanded && (
                <div className="ml-4 mt-1">
                  {group.products.map(product => {
                    const description = getProductDescription(product)
                    return (
                      <Button
                        key={product.id}
                        variant={selectedProduct?.id === product.id ? "default" : "ghost"}
                        size="sm"
                        className={cn(
                          "w-full justify-start px-2 py-1 h-auto transition-colors",
                          selectedProduct?.id === product.id && "font-semibold"
                        )}
                        onClick={() => onProductSelect(product)}
                      >
                        <div className="flex flex-col items-start w-full">
                          <div className="flex items-center gap-2 w-full">
                            {selectedProduct?.id === product.id && (
                              <span className="text-xs">▶</span>
                            )}
                            <span className="text-sm">{formatProductName(product)}</span>
                          </div>
                          {description && (
                            <span className="text-xs text-muted-foreground ml-5">
                              {description}
                            </span>
                          )}
                        </div>
                      </Button>
                    )
                  })}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}