@echo off
title Start wuchang-tongcheng
chcp 65001 >nul

echo ========================================
echo   Start wuchang-tongcheng (五常同城)
echo ========================================
echo.

cd /d "%~dp0"

echo [1/4] Cleanup ports...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1

echo.
echo [2/4] Start Go Backend (8080)...
if exist "backend\cmd\server\main.go" (
    start "Go Backend (8080)" cmd /k "cd /d %~dp0backend && go run cmd/server/main.go"
) else (
    echo WARNING: backend\cmd\server\main.go not found
)

echo.
echo [3/4] Start Admin (5173)...
if exist "frontend\package.json" (
    start "Admin (5173)" cmd /k "cd /d %~dp0frontend && npm run dev"
) else (
    echo WARNING: frontend\package.json not found
)

echo.
echo [4/4] Start PC Frontend (3000)...
if exist "frontend\pc\package.json" (
    start "PC Frontend (3000)" cmd /k "cd /d %~dp0frontend\pc && npm run dev"
) else (
    echo WARNING: frontend\pc\package.json not found
)

echo.
echo ========================================
echo   Started!
echo ========================================
echo.
echo URLs:
echo   - API:     http://localhost:8080
echo   - Admin:   http://localhost:5173
echo   - PC:      http://localhost:3000/pc
echo.
pause
