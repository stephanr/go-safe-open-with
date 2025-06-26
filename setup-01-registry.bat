@echo off
setlocal enabledelayedexpansion

ECHO.
ECHO .. Writing to Chrome Registry
ECHO .. Key: HKCU\Software\Google\Chrome\NativeMessagingHosts\com.add0n.node

REG ADD "HKCU\Software\Google\Chrome\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%LocalAPPData%\com.add0n.node\manifest-chrome.json" /f
if errorlevel 1 echo Warning: Failed to write Chrome registry entry

ECHO .. Writing to Chromium Registry
ECHO .. Key: HKCU\Software\Chromium\NativeMessagingHosts\com.add0n.node
REG ADD "HKCU\Software\Chromium\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%LocalAPPData%\com.add0n.node\manifest-chrome.json" /f
if errorlevel 1 echo Warning: Failed to write Chromium registry entry

ECHO .. Writing to Edge Registry
ECHO .. Key: HKCU\Software\Microsoft\Edge\NativeMessagingHosts\com.add0n.node
REG ADD "HKCU\Software\Microsoft\Edge\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%LocalAPPData%\com.add0n.node\manifest-chrome.json" /f
if errorlevel 1 echo Warning: Failed to write Edge registry entry

ECHO .. Writing to Firefox Registry
ECHO .. Key: HKCU\SOFTWARE\Mozilla\NativeMessagingHosts\com.add0n.node
FOR %%f in ("%LocalAPPData%") do SET SHORT_PATH=%%~sf
REG ADD "HKCU\SOFTWARE\Mozilla\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%SHORT_PATH%\com.add0n.node\manifest-firefox.json" /f
if errorlevel 1 echo Warning: Failed to write Firefox registry entry

ECHO .. Writing to Waterfox Registry
ECHO .. Key: HKCU\SOFTWARE\Waterfox\NativeMessagingHosts\com.add0n.node
REG ADD "HKCU\SOFTWARE\Waterfox\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%SHORT_PATH%\com.add0n.node\manifest-firefox.json" /f
if errorlevel 1 echo Warning: Failed to write Waterfox registry entry

ECHO .. Writing to Thunderbird Registry
ECHO .. Key: HKCU\SOFTWARE\Thunderbird\NativeMessagingHosts\com.add0n.node
REG ADD "HKCU\SOFTWARE\Thunderbird\NativeMessagingHosts\com.add0n.node" /ve /t REG_SZ /d "%SHORT_PATH%\com.add0n.node\manifest-firefox.json" /f
