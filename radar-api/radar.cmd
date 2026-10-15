@echo off
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0radar.ps1" %*
exit /b %ERRORLEVEL%
