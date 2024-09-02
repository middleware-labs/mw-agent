
set apiKey to text returned of (display dialog "Please enter the Middleware API Key (MW_API_KEY):" default answer "")
set target to text returned of (display dialog "Please enter the Middleware Target (MW_TARGET):" default answer "")

-- Write the inputs to a temporary file
do shell script "echo api-key: " & apiKey & " > /tmp/mw_agent_cfg.txt"
do shell script "echo target: " & target & " >> /tmp/mw_agent_cfg.txt"

tell application "Installer" to activate