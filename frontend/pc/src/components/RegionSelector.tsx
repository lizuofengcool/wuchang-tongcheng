'use client'

import { useEffect, useState } from 'react'
import { listRegions } from '@/lib/api'
import type { Region } from '@/lib/types'
import { useRegion } from '@/lib/region'

export default function RegionSelector() {
  const { regionId, setRegionId } = useRegion()
  const [regions, setRegions] = useState<Region[]>([])
  const [open, setOpen] = useState(false)

  useEffect(() => {
    listRegions().then(setRegions).catch(() => {})
  }, [])

  const current = regions.find((r) => r.id === regionId)

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="text-sm px-3 py-1.5 border border-gray-300 rounded hover:bg-gray-50"
      >
        📍 {current?.name || '全部地区'} ▾
      </button>
      {open && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setOpen(false)}
          />
          <div className="absolute right-0 mt-1 w-40 bg-white border border-gray-200 rounded shadow-lg z-20 max-h-64 overflow-auto">
            {regions.map((r) => (
              <button
                key={r.id}
                onClick={() => {
                  setRegionId(r.id)
                  setOpen(false)
                }}
                className={`block w-full text-left px-3 py-2 text-sm hover:bg-gray-50 ${
                  r.id === regionId ? 'text-brand-600 font-bold' : ''
                }`}
              >
                {r.name}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
