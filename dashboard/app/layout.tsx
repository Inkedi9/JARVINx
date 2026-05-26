import type { Metadata } from 'next'
import './globals.css'
import Sidebar from './components/sidebar'
import Topbar from './components/topbar'

export const metadata: Metadata = {
  title: 'JARVINx — Autonomous Runtime',
  description: 'AI Agent Runtime Dashboard',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="fr" className="dark">
      <body className="bg-bg-primary text-white min-h-screen flex">
        <Sidebar />
        <div className="flex-1 flex flex-col min-h-screen ml-[220px]">
          <Topbar />
          <main className="flex-1 p-6 overflow-auto">
            {children}
          </main>
        </div>
      </body>
    </html>
  )
}