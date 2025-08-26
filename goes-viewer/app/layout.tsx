import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { ThemeProvider } from "@/components/ui/theme-provider"

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Goes Viewer',
  description: 'GOES 16 Viewer, with images pulled from a Raspberry Pi',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <head>
        <meta name="apple-mobile-web-app-title" content="Goes Viewer" />
      </head>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
          storageKey="goes-viewer-theme"
        >
          <main className="h-full container mx-auto px-4">
            <header className="mb-2 flex justify-between items-center">
              <h1 className="text-1xl font-bold">GOES 19 Viewer</h1>
            </header>
            {children}
          </main>
        </ThemeProvider>
      </body>
    </html>
  )
}

