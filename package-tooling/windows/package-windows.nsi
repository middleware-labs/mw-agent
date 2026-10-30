;--------------------------------
; NSIS Script
; Author  - Middleware Inc
; Email   - hello@middleware.io
; Github  - https://github.com/middleware-labs
;--------------------------------

;--------------------------------
;Including Header Files
!include "FileFunc.nsh"
!include "nsDialogs.nsh"
!include "MUI2.nsh"
!include "LogicLib.nsh"
!addplugindir "Plugins"
!include "StrFunc.nsh"
!include "windows-version.nsh"
${StrRep}
;--------------------------------
;Settings
  !define APPNAME "Middleware Agent"
  !define APP_NAME_IN_INSTALLED_DIR "mw-windows-agent"
  !define CONFIG_FILE_NAME_IN_INSTALLED_DIR "agent-config.yaml"
  !define COMPANYNAME "Middleware Inc"
  !define DESCRIPTION "Middleware Agent for Microsoft Windows"
  !define DEVELOPER "Middleware Lab Inc" #License Holder

  !define BUILDNUMBER "0"
  # Files Directory
  !define BUILD_DIR "..\..\build" #Replace with the full path of install folder
  !define REPO_ROOT_DIR "..\..\"
  !define LOGO_ICON_FILE "logo.ico"
  !define MUI_ICON "logo.ico"
  !define MUI_UNICON "logo.ico"
  !define LICENSE_TEXT_FILE "${REPO_ROOT_DIR}\LICENSE"
  ;!define SPLASH_IMG_FILE "splash.bmp"
  ;!define HEADER_IMG_FILE "header.bmp"
  # These three must be integers
  # These will be displayed by the "Click here for support information" link in "Add/Remove Programs"
  # It is possible to use "mailto:" links in here to open email client
  !define HELPURL "https://middleware.io/contact-us/"
  !define ABOUTURL "https://middleware.io/about-us/"
  # This is the size (in kB) of all the files copied into "Program Files"
  ;!define INSTALLSIZE 1118721

  # Set compression method
  SetCompressor lzma
;--------------------------------
;General

  ;Name and file
  Name "${APPNAME}"
  Icon "logo.ico"
  OutFile "mw-windows-agent-${VERSION}-setup.exe"

  ;Default installation folder
  InstallDir "$PROGRAMFILES64\${APPNAME}"

  ;Get installation folder from registry if available
  InstallDirRegKey HKLM "Software\${APPNAME}" ""

  ;Request application privileges for Windows Vista
  RequestExecutionLevel admin ;Require admin rights on NT6+ (When UAC is turned on)

;--------------------------------
;Variables

Var StartMenuFolder

Var InstallerStatus
Var InstallerMessage
Var InstallerOperation
Var InstallServiceStatus
Var StartServiceStatus
  
;--------------------------------
;Interface Settings

;!define MUI_HEADERIMAGE
;!define MUI_HEADERIMAGE_BITMAP "${HEADER_IMG_FILE}" ; optional
  
!define MUI_ABORTWARNING

;--------------------------------
;Pages
  !insertmacro MUI_PAGE_WELCOME
  !insertmacro MUI_PAGE_LICENSE "${LICENSE_TEXT_FILE}"
  Page custom pgPageCreate pgPageLeave
  !insertmacro MUI_PAGE_DIRECTORY

  ;Start Menu Folder Page Configuration
  !define MUI_STARTMENUPAGE_REGISTRY_ROOT "HKLM" 
  !define MUI_STARTMENUPAGE_REGISTRY_KEY "Software\${APPNAME}" 
  !define MUI_STARTMENUPAGE_REGISTRY_VALUENAME "Start Menu Folder"
  !insertmacro MUI_PAGE_STARTMENU Application $StartMenuFolder

  !insertmacro MUI_PAGE_INSTFILES
  !insertmacro MUI_PAGE_FINISH
  !insertmacro MUI_UNPAGE_WELCOME
  !insertmacro MUI_UNPAGE_CONFIRM
  !insertmacro MUI_UNPAGE_INSTFILES
  !insertmacro MUI_UNPAGE_FINISH

;--------------------------------
;Languages

  !insertmacro MUI_LANGUAGE "English"

;--------------------------------
;Verify User is Admin or not

  !macro VerifyUserIsAdmin
  UserInfo::GetAccountType
  pop $0
  ${If} $0 != "admin" ;Require admin rights on NT4+
	messageBox mb_iconstop "Administrator rights required!"
	setErrorLevel 740 ;ERROR_ELEVATION_REQUIRED
	quit
  ${EndIf}
  !macroend


;--------------------------------
Var Dialog
Var TextAPIKey
Var TextTarget
Var MWAPIKey
Var MWTarget

# Optional agent settings. These mirror the MW_* variables the Linux packages
# accept and map to the flags defined in cmd/host-agent/main.go.
Var MWConfigCheckInterval
Var MWHostTags
Var MWAgentInternalMetricsPort
Var MWAPIURLForConfigCheck
Var MWSyntheticMonitoringAPIURL
Var MWFetchAccountOtelConfig
Var MWFeatureInfraMonitoring
Var MWFeatureMetricCollection
Var MWFeatureLogCollection
Var MWFeatureSyntheticMonitoring
Var MWFeatureServiceReporting
Var MWServiceReportInterval
Var MWAgentSelfProfiling
Var MWProfilingServerURL
Var MWLogFile
Var MWLogFileSize
Var MWLoggingLevel
Var MWGRPCPort
Var MWHTTPPort
Var MWFluentPort
Var MWDockerEndpoint
Var MWEnableDatadogReceiver

# Read /NAME=value from the installer command line into VAR ("" if absent).
# Values may be wrapped in "", '' or `` quotes.
!macro ReadOption NAME VAR
  ${GetOptions} "$0" "/${NAME}=" ${VAR}
!macroend

# Bool flags must be written as unquoted true/false or the agent fails to
# parse its config, so anything else is dropped and the agent default applies.
!macro NormalizeBool VAR
  ${If} ${VAR} == "true"
    StrCpy ${VAR} "true"
  ${ElseIf} ${VAR} == "false"
    StrCpy ${VAR} "false"
  ${Else}
    StrCpy ${VAR} ""
  ${EndIf}
!macroend

# Same for numeric flags: drop values that are not plain integers.
!macro NormalizeInt VAR
  IntOp $R1 ${VAR} + 0
  ${If} "$R1" != "${VAR}"
  ${OrIf} $R1 < 0
    StrCpy ${VAR} ""
  ${EndIf}
!macroend

Function .onInit
  Strcpy $InstallerOperation "install"
  StrCpy $InstallerStatus "pending"
  StrCpy $InstallerMessage "installer loaded"
  StrCpy $InstallServiceStatus "not yet installed"
  StrCpy $StartServiceStatus "not yet started"

  # Get input arguments if passed
  ${GetParameters} $0
  !insertmacro ReadOption "MW_API_KEY" $MWAPIKey
  !insertmacro ReadOption "MW_TARGET" $MWTarget
  !insertmacro ReadOption "MW_CONFIG_CHECK_INTERVAL" $MWConfigCheckInterval
  !insertmacro ReadOption "MW_HOST_TAGS" $MWHostTags
  !insertmacro ReadOption "MW_AGENT_INTERNAL_METRICS_PORT" $MWAgentInternalMetricsPort
  !insertmacro ReadOption "MW_API_URL_FOR_CONFIG_CHECK" $MWAPIURLForConfigCheck
  !insertmacro ReadOption "MW_SYNTHETIC_MONITORING_API_URL" $MWSyntheticMonitoringAPIURL
  ${If} $MWSyntheticMonitoringAPIURL == ""
    !insertmacro ReadOption "MW_API_URL_FOR_SYNTHETIC_MONITORING" $MWSyntheticMonitoringAPIURL
  ${EndIf}
  !insertmacro ReadOption "MW_FETCH_ACCOUNT_OTEL_CONFIG" $MWFetchAccountOtelConfig
  !insertmacro ReadOption "MW_AGENT_FEATURES_INFRA_MONITORING" $MWFeatureInfraMonitoring
  !insertmacro ReadOption "MW_AGENT_FEATURES_METRIC_COLLECTION" $MWFeatureMetricCollection
  !insertmacro ReadOption "MW_AGENT_FEATURES_LOG_COLLECTION" $MWFeatureLogCollection
  !insertmacro ReadOption "MW_AGENT_FEATURES_SYNTHETIC_MONITORING" $MWFeatureSyntheticMonitoring
  !insertmacro ReadOption "MW_AGENT_FEATURES_REPORT_SERVICES" $MWFeatureServiceReporting
  !insertmacro ReadOption "MW_SERVICE_REPORT_INTERVAL" $MWServiceReportInterval
  !insertmacro ReadOption "MW_AGENT_SELF_PROFILING" $MWAgentSelfProfiling
  !insertmacro ReadOption "MW_PROFILING_SERVER_URL" $MWProfilingServerURL
  !insertmacro ReadOption "MW_LOGFILE" $MWLogFile
  !insertmacro ReadOption "MW_LOGFILE_SIZE" $MWLogFileSize
  !insertmacro ReadOption "MW_LOGGING_LEVEL" $MWLoggingLevel
  !insertmacro ReadOption "MW_AGENT_GRPC_PORT" $MWGRPCPort
  !insertmacro ReadOption "MW_AGENT_HTTP_PORT" $MWHTTPPort
  !insertmacro ReadOption "MW_AGENT_FLUENT_PORT" $MWFluentPort
  ${If} $MWFluentPort == ""
    !insertmacro ReadOption "MW_FLUENT_PORT" $MWFluentPort
  ${EndIf}
  !insertmacro ReadOption "MW_DOCKER_ENDPOINT" $MWDockerEndpoint
  !insertmacro ReadOption "MW_ENABLE_DATADOG_RECEIVER" $MWEnableDatadogReceiver

  !insertmacro NormalizeBool $MWFetchAccountOtelConfig
  !insertmacro NormalizeBool $MWFeatureInfraMonitoring
  !insertmacro NormalizeBool $MWFeatureMetricCollection
  !insertmacro NormalizeBool $MWFeatureLogCollection
  !insertmacro NormalizeBool $MWFeatureSyntheticMonitoring
  !insertmacro NormalizeBool $MWFeatureServiceReporting
  !insertmacro NormalizeBool $MWAgentSelfProfiling
  !insertmacro NormalizeBool $MWEnableDatadogReceiver
  !insertmacro NormalizeInt $MWAgentInternalMetricsPort
  !insertmacro NormalizeInt $MWLogFileSize
FunctionEnd

Function onManualInstallClick
    pop $R9
    ExecShell "open" "https://app.middleware.io" 
FunctionEnd

Function pgPageCreate
    !insertmacro MUI_HEADER_TEXT "Middleware Settings" "Please provide API Key and Target URL for your Middleware account"

    nsDialogs::Create 1018
    Pop $Dialog

    ${If} $Dialog == error
        Abort
    ${EndIf}

    ${NSD_CreateGroupBox} 5% 10u 90% 70u "Middleware Account Details"
    Pop $0

        ${NSD_CreateLabel} 10% 26u 30% 13u "API Key (MW_API_KEY)     :"
        Pop $0

        ${NSD_CreateText} 40% 24u 50% 14u "$MWAPIKey"
        Pop $TextAPIKey

        ${NSD_CreateLabel} 10% 55u 30% 13u "Target URL (MW_TARGET) :"
        Pop $0

        ${NSD_CreateText} 40% 53u 50% 14u "$MWTarget"
        Pop $TextTarget
    
    ${NSD_CreateLabel} 5% 86u 90% 34u "API Key and Target URL can be found in the installation section of your Middleware account."
    Pop $0

    ${NSD_CreateLink}  5% 120u 90% 34u "Click here to access your API Key and Target URL."
    Pop $R9

    ${NSD_OnClick} $R9 onManualInstallClick

    nsDialogs::Show
FunctionEnd

Function PgPageLeave
    ${NSD_GetText} $TextAPIKey $MWAPIKey
    ${NSD_GetText} $TextTarget $MWTarget
    ${If} $MWAPIKey == ""
        MessageBox MB_OK "Please provide a valid Middleware API key"
        Abort
    ${EndIf}
    ${If} $MWTarget == ""
        MessageBox MB_OK "Please provide a valid Middleware target"
        Abort
    ${EndIf}
    StrCpy $InstallerMessage "api key and target entered"
FunctionEnd

# Write KEY: "VALUE" to the open config file ($4), escaping backslashes and
# double quotes for YAML. Always writes, even when VALUE is empty.
!macro WriteStringSetting INDENT KEY VALUE
  ${StrRep} $R0 "${VALUE}" "\" "\\"
  ${StrRep} $R0 $R0 '"' '\"'
  FileWrite $4 '${INDENT}${KEY}: "$R0"$\r$\n'
!macroend

# Write KEY: "VALUE" only when VALUE is set.
!macro WriteOptionalString INDENT KEY VALUE
  ${If} "${VALUE}" != ""
    !insertmacro WriteStringSetting "${INDENT}" "${KEY}" "${VALUE}"
  ${EndIf}
!macroend

# Write KEY: VALUE unquoted (bool and int flags) only when VALUE is set.
!macro WriteOptionalRaw INDENT KEY VALUE
  ${If} "${VALUE}" != ""
    FileWrite $4 "${INDENT}${KEY}: ${VALUE}$\r$\n"
  ${EndIf}
!macroend

Function UpdateConfigFile
    ${If} $MWConfigCheckInterval == ""
        StrCpy $MWConfigCheckInterval "5m"
    ${EndIf}
    ${If} $MWLogFile == ""
        StrCpy $MWLogFile "$INSTDIR\mw-agent.log"
    ${EndIf}

    FileOpen $4 "$INSTDIR\${CONFIG_FILE_NAME_IN_INSTALLED_DIR}" w
    !insertmacro WriteStringSetting "" "api-key" $MWAPIKey
    !insertmacro WriteStringSetting "" "target" $MWTarget
    !insertmacro WriteStringSetting "" "config-check-interval" $MWConfigCheckInterval
    !insertmacro WriteStringSetting "" "logfile" $MWLogFile
    !insertmacro WriteOptionalRaw "" "logfile-size" $MWLogFileSize
    !insertmacro WriteOptionalString "" "logging-level" $MWLoggingLevel
    !insertmacro WriteOptionalString "" "host-tags" $MWHostTags
    !insertmacro WriteOptionalRaw "" "agent-internal-metrics-port" $MWAgentInternalMetricsPort
    !insertmacro WriteOptionalString "" "api-url-for-config-check" $MWAPIURLForConfigCheck
    !insertmacro WriteOptionalRaw "" "fetch-account-otel-config" $MWFetchAccountOtelConfig
    !insertmacro WriteOptionalString "" "service-report-interval" $MWServiceReportInterval
    !insertmacro WriteOptionalRaw "" "agent-self-profiling" $MWAgentSelfProfiling
    !insertmacro WriteOptionalString "" "profiling-server-url" $MWProfilingServerURL
    !insertmacro WriteOptionalString "" "grpc-port" $MWGRPCPort
    !insertmacro WriteOptionalString "" "http-port" $MWHTTPPort
    !insertmacro WriteOptionalString "" "fluent-port" $MWFluentPort
    !insertmacro WriteOptionalString "" "docker-endpoint" $MWDockerEndpoint
    !insertmacro WriteOptionalRaw "" "enable-datadog-receiver" $MWEnableDatadogReceiver

    ${If} "$MWFeatureInfraMonitoring$MWFeatureMetricCollection$MWFeatureLogCollection$MWFeatureSyntheticMonitoring$MWFeatureServiceReporting" != ""
        FileWrite $4 "agent-features:$\r$\n"
        !insertmacro WriteOptionalRaw "  " "infra-monitoring" $MWFeatureInfraMonitoring
        !insertmacro WriteOptionalRaw "  " "metric-collection" $MWFeatureMetricCollection
        !insertmacro WriteOptionalRaw "  " "log-collection" $MWFeatureLogCollection
        !insertmacro WriteOptionalRaw "  " "synthetic-monitoring" $MWFeatureSyntheticMonitoring
        !insertmacro WriteOptionalRaw "  " "service-reporting" $MWFeatureServiceReporting
    ${EndIf}

    ${If} $MWSyntheticMonitoringAPIURL != ""
        FileWrite $4 "synthetic-monitoring:$\r$\n"
        !insertmacro WriteStringSetting "  " "api-url" $MWSyntheticMonitoringAPIURL
    ${EndIf}
    FileClose $4
    StrCpy $InstallerMessage "$INSTDIR\${CONFIG_FILE_NAME_IN_INSTALLED_DIR} file created"
FunctionEnd
;--------------------------------
;Installer section

Section "install"
  # Files for install directory - to build the installer, these should be in the same directory as the install script (this file)
  SetOutPath $INSTDIR

  ################################################################################################################
    
  #Add your Files Here
  # Files add here should be removed by the uninstaller (see section "uninstall")
  file "${BUILD_DIR}\${APP_NAME_IN_INSTALLED_DIR}.exe"
  file "logo.ico"
  file /oname=${CONFIG_FILE_NAME_IN_INSTALLED_DIR} "..\${CONFIG_FILE_NAME_IN_INSTALLED_DIR}.sample"

  Call UpdateConfigFile
 
  SimpleSC::InstallService ${APP_NAME_IN_INSTALLED_DIR} "Middleware Agent" "16" "2" "$\"$INSTDIR\${APP_NAME_IN_INSTALLED_DIR}.exe$\" start --config-file $\"$INSTDIR\${CONFIG_FILE_NAME_IN_INSTALLED_DIR}$\"" "" "" ""
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)
  StrCpy $InstallServiceStatus $0

  SimpleSC::StartService "${APP_NAME_IN_INSTALLED_DIR}" "" 30
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)
  StrCpy $StartServiceStatus $0
  StrCpy $InstallerStatus "installed"
  StrCpy $InstallerMessage "installer completed"
 
  ################################################################################################################

  # Uninstaller - see function un.onInit and section "uninstall" for configuration
  writeUninstaller "$INSTDIR\uninstall.exe"

  SetOutPath $INSTDIR
  # Start Menu
  CreateDirectory "$SMPROGRAMS\${APPNAME}"
  CreateShortCut "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk" "$INSTDIR\${APP_NAME_IN_INSTALLED_DIR}.exe" "" "$INSTDIR\logo.ico"
  CreateShortCut "$SMPROGRAMS\${APPNAME}\uninstall.lnk" "$INSTDIR\uninstall.exe" "" ""
  
  # Desktop Shortcut
;  CreateShortCut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\${APP_NAME_IN_INSTALLED_DIR}.exe" "" "$INSTDIR\logo.ico"

  # Registry information for add/remove programs
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME} - ${DESCRIPTION}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "QuitUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$\"$INSTDIR\logo.ico$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "$\"${COMPANYNAME}$\""
 ; WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "HelpLink" "$\"${HELPURL}$\""
 ; WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion;\Uninstall\${APPNAME}" "URLUpdateInfo" "$\"${UPDATEURL}$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "URLInfoAbout" "$\"${ABOUTURL}$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "$\"${VERSION}.${BUILDNUMBER}$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Version" ${VERSION}.${BUILDNUMBER}
  # There is no option for modifying or reparing the install
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
  # Set the INSTALLSIZE constant (!define at the top of this script) so Add/Remove Program can accurately report the size
;  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "EstimatedSize" ${INSTALLSIZE}
SectionEnd

;--------------------------------
;Version Information

  VIProductVersion "${VERSION}.${BUILDNUMBER}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "ProductName" "${APPNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "Comments" "${DESCRIPTION}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "CompanyName" "${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "LegalTrademarks" "${APPNAME} is a trademark of ${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "LegalCopyright" "${DEVELOPER} | ${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "FileDescription" "${APPNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "FileVersion" "${VERSION}.${BUILDNUMBER}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "ProductVersion" "${VERSION}.${BUILDNUMBER}"

;--------------------------------
;Verify Uninstall

  function un.onInit		
	# Verify the uninstaller - last chance to back out
	MessageBox MB_OKCANCEL "Permanantly remove ${APPNAME}?" /SD IDOK IDOK next
		Abort
	next:
	!insertmacro VerifyUserIsAdmin
  functionEnd
  
;--------------------------------
;Uninstaller Section

Section "uninstall"

  #Remove Start Menu Launcher
  delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
  delete "$SMPROGRAMS\${APPNAME}\uninstall.lnk"
  #Remove Desktop Shortcut
  delete "$DESKTOP\${APPNAME}.lnk"
  #Try to remove the Start Menu folder - this will only happen if it is empty
  rmDir "$SMPROGRAMS\${APPNAME}"

  ;ExecWait "sc delete ${APP_NAME_IN_INSTALLED_DIR}"
  SimpleSC::StopService "${APP_NAME_IN_INSTALLED_DIR}" 1 30
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)
 
  SimpleSC::RemoveService "${APP_NAME_IN_INSTALLED_DIR}"
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)

  ################################################################################################################
  #Delete installation folder from registry if available - this will only happen if it is empty
  rmDir /r "$INSTDIR"

  DeleteRegKey /ifempty HKLM "Software\${APPNAME}"

  # Remove uninstaller information from the registry
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"
SectionEnd

Function .onGUIEnd
  # your code here
   ReadRegStr $0 HKLM "System\CurrentControlSet\Control\ComputerName\ActiveComputerName" "ComputerName"
   ${GetWindowsVersion} $R0
   Var /GLOBAL data 
   StrCpy $data "\
   {\
    $\"status$\": $\"$InstallerStatus$\",\
    $\"metadata$\": {\
      $\"script$\": $\"windows$\",\
      $\"message$\": $\"$InstallerMessage$\",\
      $\"operation$\": $\"$InstallerOperation$\",\
      $\"windows_version$\": $\"$R0$\",\
      $\"agent_version$\": $\"${VERSION}.${BUILDNUMBER}$\",\
      $\"host_id$\": $\"$0$\",\
      $\"script_logs$\": $\"agent_service_install_status: $InstallServiceStatus, agent_service_start_status: $StartServiceStatus$\"\
    }\
  }"

  NScurl::http POST "$MWTarget/api/v1/agent/tracking/$MWAPIKey" MEMORY /HEADER "Content-Type: application/json" /DATA '$data' /END
  Pop $0

FunctionEnd
;--------------------------------
