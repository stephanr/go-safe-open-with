mkdir %LOCALAPPDATA%\com.add0n.node
dir %LOCALAPPDATA%\com.add0n.node

copy /V /Y manifest-chrome.json %LOCALAPPDATA%\com.add0n.node\manifest-chrome.json
copy /V /Y manifest-firefox.json %LOCALAPPDATA%\com.add0n.node\manifest-firefox.json