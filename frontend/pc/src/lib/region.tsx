// 客户端地区上下文（React Context）
// PC门户默认展示一个地区（默认武汉市=2，见后端 seed），用户可切换

'use client'

import { createContext, useContext, useState, ReactNode } from 'react'

interface RegionContextValue {
  regionId: number
  setRegionId: (id: number) => void
}

const RegionContext = createContext<RegionContextValue | null>(null)

const DEFAULT_REGION_ID = Number(process.env.NEXT_PUBLIC_DEFAULT_REGION_ID) || 2

export function RegionProvider({ children }: { children: ReactNode }) {
  const [regionId, setRegionId] = useState<number>(DEFAULT_REGION_ID)
  return (
    <RegionContext.Provider value={{ regionId, setRegionId }}>
      {children}
    </RegionContext.Provider>
  )
}

export function useRegion() {
  const ctx = useContext(RegionContext)
  if (!ctx) throw new Error('useRegion must be used within RegionProvider')
  return ctx
}
