@echo off
title Stop wuchang-tongcheng
chcp 65001 >nul

echo ========================================
echo   Stop wuchang-tongcheng (五常同城)
echo ========================================
echo.

echo Stopping services on ports 8080, 5173, 3000...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8080" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":5173" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":3000" ^| findstr LISTENING') do taskkill /F /PID %%a >nul 2>&1

echo.
echo ========================================
echo   Stopped!
echo ========================================
echo.
pause
