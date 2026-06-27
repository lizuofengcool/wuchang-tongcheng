import type { Metadata } from 'next'
import './globals.css'
import Header from '@/components/Header'
import Footer from '@/components/Footer'
import { RegionProvider } from '@/lib/region'

export const metadata: Metadata = {
  title: '五常同城 - 本地生活服务平台',
  description: '面向五常市的本地生活服务平台，提供分类信息、同城头条、商家服务。',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="zh-CN">
      <body>
        <RegionProvider>
          <div className="min-h-screen flex flex-col">
            <Header />
            <main className="flex-1">{children}</main>
            <Footer />
          </div>
        </RegionProvider>
      </body>
    </html>
  )
}
