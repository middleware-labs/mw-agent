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
Var LogFilePath

Function .onInit
  Strcpy $InstallerOperation "install"
  StrCpy $InstallerStatus "pending"
  StrCpy $InstallerMessage "installer loaded"
  StrCpy $InstallServiceStatus "not yet installed"
  StrCpy $StartServiceStatus "not yet started"

  # Get input arguments if passed
  ${GetParameters} $0
  ${GetOptions} "$0" "/MW_API_KEY=" $1
  ${GetOptions} "$0" "/MW_TARGET=" $2

  StrCpy $MWAPIKey $1
  StrCpy $MWTarget $2
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

Function UpdateConfigFile
    FileOpen $4 "$INSTDIR\${CONFIG_FILE_NAME_IN_INSTALLED_DIR}" w
    FileWrite $4 "api-key: $MWAPIKey $\r$\n"
    FileWrite $4 "target: $MWTarget $\r$\n"
    FileWrite $4 "config-check-interval: $\"5m$\"$\r$\n"
    Strcpy $LogFilePath "$INSTDIR\mw-agent.log"
    ${StrRep} $R0 $LogFilePath "\" "\\"
    FileWrite $4 "logfile: $\"$R0$\"$\r$\n"
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
  file "${REPO_ROOT_DIR}\${CONFIG_FILE_NAME_IN_INSTALLED_DIR}"

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
	MessageBox MB_OKCANCEL "Permanantly remove ${APPNAME}?" IDOK next
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
