@echo off
setlocal

cd /d "%~dp0"

echo Building GhostDrop for Windows...

if not exist "ghostdrop-windows" (
    mkdir "ghostdrop-windows"
)

set GOOS=windows
set GOARCH=amd64

go build ^
    -trimpath ^
    -ldflags="-s -w" ^
    -o "ghostdrop-windows\ghostdrop.exe" ^
    ./cmd/ghostdrop

if errorlevel 1 (
    echo.
    echo Build failed.
    exit /b 1
)

echo.
echo Build completed:
echo   ghostdrop-windows\ghostdrop.exe

endlocal
