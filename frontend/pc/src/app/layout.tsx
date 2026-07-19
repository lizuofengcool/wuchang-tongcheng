import type { Metadata } from 'next'
import './globals.css'
import Header from '@/components/Header'
import Footer from '@/components/Footer'
import { RegionProvider } from '@/lib/region'

export const metadata: Metadata = {
  title: {
    default: '五常同城 - 本地生活服务平台',
    template: '%s | 五常同城',
  },
  description:
    '五常同城是面向五常市的本地生活服务平台，提供二手交易、招聘求职、房屋租售、同城头条、商家服务、同城商城等分类信息。',
  keywords: [
    '五常同城',
    '五常信息港',
    '五常二手',
    '五常招聘',
    '五常房产',
    '五常黄页',
    '五常商城',
    '同城头条',
    '本地生活服务',
  ],
  authors: [{ name: '五常同城' }],
  metadataBase: new URL('http://localhost:3010'),
  openGraph: {
    title: '五常同城 - 本地生活服务平台',
    description:
      '面向五常市的本地生活服务平台，提供二手、招聘、房产、黄页、商城、头条等分类信息。',
    siteName: '五常同城',
    type: 'website',
    locale: 'zh_CN',
  },
  robots: {
    index: true,
    follow: true,
  },
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
