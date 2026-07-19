// shadcn/ui 必需的 cn 工具函数
// 依赖 clsx + tailwind-merge（已加入 package.json devDependencies）
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
