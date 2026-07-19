@echo off
setlocal EnableDelayedExpansion
title Stop wuchang-tongcheng
chcp 65001 >nul

echo ========================================
echo   Stop wuchang-tongcheng (五常同城)
echo ========================================
echo.

cd /d "%~dp0"

echo Loading ports from .env ...
for /f "usebackq tokens=1,2 delims==" %%a in (".env") do (
    set "key=%%a"
    if not "!key!"=="" if not "!key:~0,1!"=="#" (
        set "%%a=%%b"
    )
)

if "%WCTC_SERVER_PORT%"=="" set "WCTC_SERVER_PORT=8088"
if "%WCTC_ADMIN_PORT%"=="" set "WCTC_ADMIN_PORT=5177"
if "%WCTC_PC_PORT%"=="" set "WCTC_PC_PORT=3010"

echo Stopping services on ports %WCTC_SERVER_PORT%, %WCTC_ADMIN_PORT%, %WCTC_PC_PORT%...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_SERVER_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_ADMIN_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":%WCTC_PC_PORT%" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1

echo.
echo ========================================
echo   Stopped!
echo ========================================
echo.
pause
