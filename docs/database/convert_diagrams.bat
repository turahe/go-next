@echo off
echo Converting Mermaid diagrams to SVG using npx...
echo.

REM Check if Node.js is installed
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js is not installed. Please install Node.js first.
    echo Download from: https://nodejs.org/
    pause
    exit /b 1
)

echo ✅ Node.js is installed

REM Convert the main ERD diagram
if exist erd.mmd (
    echo 🔄 Converting erd.mmd to erd.svg...
    npx @mermaid-js/mermaid-cli -i erd.mmd -o erd.svg
    if errorlevel 1 (
        echo ❌ Failed to convert erd.mmd
    ) else (
        echo ✅ Successfully converted erd.mmd to erd.svg
    )
) else (
    echo ❌ erd.mmd not found
)

REM Convert any other .mmd files in the directory
for %%f in (*.mmd) do (
    if not "%%f"=="erd.mmd" (
        echo 🔄 Converting %%f to %%~nf.svg...
        npx @mermaid-js/mermaid-cli -i "%%f" -o "%%~nf.svg"
        if errorlevel 1 (
            echo ❌ Failed to convert %%f
        ) else (
            echo ✅ Successfully converted %%f to %%~nf.svg
        )
    )
)

echo.
echo 🎉 Conversion complete!
echo 📁 SVG files saved in current directory
pause 