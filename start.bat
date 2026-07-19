@echo off
setlocal EnableDelayedExpansion
title Start wuchang-tongcheng
chcp 65001 >nul

echo ========================================
echo   Start wuchang-tongcheng (五常同城)
echo ========================================
echo.

cd /d "%~dp0"

echo [0/4] Loading ports from .env ...
for /f "usebackq tokens=1,2 delims==" %%a in (".env") do (
    set "key=%%a"
    if not "!key!"=="" if not "!key:~0,1!"=="#" (
        set "%%a=%%b"
    )
)

if "%WCTC_SERVER_PORT%"=="" set "WCTC_SERVER_PORT=8088"
if "%WCTC_ADMIN_PORT%"=="" set "WCTC_ADMIN_PORT=5177"
if "%WCTC_PC_PORT%"=="" set "WCTC_PC_PORT=3010"

echo   SERVER_PORT=%WCTC_SERVER_PORT%
echo   ADMIN_PORT =%WCTC_ADMIN_PORT%
echo   PC_PORT    =%WCTC_PC_PORT%
echo.

echo [1/4] Cleanup ports...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_SERVER_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_ADMIN_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_PC_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1

echo.
echo [2/4] Start Go Backend (%WCTC_SERVER_PORT%)...
if exist "backend\cmd\server\main.go" (
    start "Go Backend (%WCTC_SERVER_PORT%)" cmd /k "cd /d %~dp0backend && go run cmd/server/main.go"
) else (
    echo WARNING: backend\cmd\server\main.go not found
)

echo.
echo [3/4] Start Admin (%WCTC_ADMIN_PORT%)...
if exist "frontend\package.json" (
    start "Admin (%WCTC_ADMIN_PORT%)" cmd /k "cd /d %~dp0frontend && set VITE_PORT=%WCTC_ADMIN_PORT% && npm run dev"
) else (
    echo WARNING: frontend\package.json not found
)

echo.
echo [4/4] Start PC Frontend (%WCTC_PC_PORT%)...
if exist "frontend\pc\package.json" (
    start "PC Frontend (%WCTC_PC_PORT%)" cmd /k "cd /d %~dp0frontend\pc && set PORT=%WCTC_PC_PORT% && npm run dev"
) else (
    echo WARNING: frontend\pc\package.json not found
)

echo.
echo ========================================
echo   Started!
echo ========================================
echo.
echo URLs:
echo   - API:     http://localhost:%WCTC_SERVER_PORT%
echo   - Admin:   http://localhost:%WCTC_ADMIN_PORT%
echo   - PC:      http://localhost:%WCTC_PC_PORT%/pc
echo.
pause
