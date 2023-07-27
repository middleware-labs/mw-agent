;--------------------------------
; NSIS Script
; Author  - Middleware Inc
; Email   - hello@middleware.io
; Github  - https://github.com/middleware-labs
;--------------------------------

;--------------------------------
;Including Header Files

!include "nsDialogs.nsh"
!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "ReplaceInFile3.nsh"
!addplugindir "Plugins"
  


;--------------------------------
;Settings
  !define APPNAME "Middleware Windows Agent"
  !define APP_NAME_IN_INSTALLED_DIR "mw-windows-agent"
  !define CONFIG_FILE_NAME_IN_INSTALLED_DIR "config.yaml"
  !define COMPANYNAME "Middleware Inc"
  !define DESCRIPTION "Middleware Windows Agent."
  !define DEVELOPER "Middleware Inc" #License Holder
  # Files Directory
  ;!define FILE_DIR "windows" #Replace with the full path of install folder
  !define LOGO_ICON_FILE "logo.ico"
  !define MUI_ICON "logo.ico"
  !define MUI_UNICON "logo.ico"
  !define LICENSE_TEXT_FILE "LICENSE.txt"
  ;!define SPLASH_IMG_FILE "splash.bmp"
  ;!define HEADER_IMG_FILE "header.bmp"
  # These three must be integers
  !define VERSIONMAJOR 1	#Major release Number
  !define VERSIONMINOR 1	#Minor release Number
  !define VERSIONBUILD 1	#Maintenance release Number (bugfixes only)
  !define BUILDNUMBER 1		#Source control revision number
  # These will be displayed by the "Click here for support information" link in "Add/Remove Programs"
  # It is possible to use "mailto:" links in here to open email client
  !define HELPURL "https://middleware.io/contact-us/"
  !define ABOUTURL "https://middleware.io/about-us/"
  # This is the size (in kB) of all the files copied into "Program Files"
  ;!define INSTALLSIZE 1118721

;--------------------------------
;General

  ;Name and file
  Name "${APPNAME}"
  Icon "logo.ico"
  OutFile "mw-windows-agent-setup.exe"

  ;Default installation folder
  InstallDir "$PROGRAMFILES\${APPNAME}"

  ;Get installation folder from registry if available
  InstallDirRegKey HKCU "Software\${APPNAME}" ""

  ;Request application privileges for Windows Vista
  RequestExecutionLevel admin ;Require admin rights on NT6+ (When UAC is turned on)

;--------------------------------
;Variables

Var StartMenuFolder
  
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
  ;!define MUI_STARTMENUPAGE_REGISTRY_ROOT "HKCU" 
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

Function pgPageCreate
    !insertmacro MUI_HEADER_TEXT "Middleware Settings" "Please provide API Key and Target URL for your Middleware account"

    nsDialogs::Create 1018
    Pop $Dialog

    ${If} $Dialog == error
        Abort
    ${EndIf}

    ${NSD_CreateGroupBox} 5% 10u 90% 70u "Middleware Account Details"
    Pop $0

        ${NSD_CreateLabel} 10% 26u 30% 14u "API Key (MW_API_KEY) :"
        Pop $0

        ${NSD_CreateText} 40% 24u 50% 14u ""
        Pop $TextAPIKey

        ${NSD_CreateLabel} 10% 55u 30% 14u "Target (MW_TARGET)   :"
        Pop $0

        ${NSD_CreateText} 40% 53u 50% 14u ""
        Pop $TextTarget
    
    ${NSD_CreateLabel} 5% 86u 90% 34u "API key and Target can be found in the installation section of your Middleware account."
    Pop $0
    nsDialogs::Show
FunctionEnd

Function PgPageLeave
    ${NSD_GetText} $TextAPIKey $0
    ${NSD_GetText} $TextTarget $1
    ${If} $0 == ""
        MessageBox MB_OK "Please provide a valid Middleware API key"
        Abort
    ${EndIf}
    ${If} $1 == ""
        MessageBox MB_OK "Please provide a valid Middleware target"
        Abort
    ${EndIf}
FunctionEnd

Function UpdateConfigFile
    StrCpy $OLD_STR "api-key:"
    StrCpy $FST_OCC all
    StrCpy $NR_OCC all
    StrCpy $REPLACEMENT_STR "api-key: $0"
    StrCpy $FILE_TO_MODIFIED "$INSTDIR\config.yaml"
    !insertmacro ReplaceInFile $OLD_STR $FST_OCC $NR_OCC $REPLACEMENT_STR $FILE_TO_MODIFIED

    StrCpy $OLD_STR 'target:'
    StrCpy $FST_OCC all
    StrCpy $NR_OCC all
    StrCpy $REPLACEMENT_STR "target: $1"
    !insertmacro ReplaceInFile $OLD_STR $FST_OCC $NR_OCC $REPLACEMENT_STR $FILE_TO_MODIFIED


FunctionEnd
;--------------------------------
;Installer section

Section "install"
  # Files for install directory - to build the installer, these should be in the same directory as the install script (this file)
  SetOutPath $INSTDIR

  ################################################################################################################
    
  #Add your Files Here
  # Files add here should be removed by the uninstaller (see section "uninstall")
  file "${APP_NAME_IN_INSTALLED_DIR}.exe"
  file "logo.ico"
  file "${CONFIG_FILE_NAME_IN_INSTALLED_DIR}"
  file /r "configyamls"

  Call UpdateConfigFile
  ;ExecWait 'sc create ${APP_NAME_IN_INSTALLED_DIR} error= "severe" displayname= "${APPNAME}" type= "own" start= "auto" binpath= "$INSTDIR\${APP_NAME_IN_INSTALLED_DIR}.exe start --config-file $INSTDIR\config.yaml"'
  ;ExecWait 'net start ${APP_NAME_IN_INSTALLED_DIR}'
  SimpleSC::InstallService ${APP_NAME_IN_INSTALLED_DIR} "Middleware Windows Agent" "16" "2" "$\"$INSTDIR\${APP_NAME_IN_INSTALLED_DIR}.exe$\" start --config-file $\"$INSTDIR\config.yaml$\"" "" "" ""
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)

  SimpleSC::StartService "${APP_NAME_IN_INSTALLED_DIR}" "" 30
  Pop $0 ; returns an errorcode (<>0) otherwise success (0)
 
 
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
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME} - ${DESCRIPTION}"
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "QuitUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$\"$INSTDIR\logo.ico$\""
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "$\"${COMPANYNAME}$\""
 ; WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "HelpLink" "$\"${HELPURL}$\""
 ; WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion;\Uninstall\${APPNAME}" "URLUpdateInfo" "$\"${UPDATEURL}$\""
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "URLInfoAbout" "$\"${ABOUTURL}$\""
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "$\"${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.${BUILDNUMBER}$\""
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMajor" ${VERSIONMAJOR}
  WriteRegStr HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMinor" ${VERSIONMINOR}
  # There is no option for modifying or reparing the install
  WriteRegDWORD HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
  # Set the INSTALLSIZE constant (!define at the top of this script) so Add/Remove Program can accurately report the size
;  WriteRegDWORD HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}" "EstimatedSize" ${INSTALLSIZE}
SectionEnd

;--------------------------------
;Version Information

  VIProductVersion "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.${BUILDNUMBER}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "ProductName" "${APPNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "Comments" "${DESCRIPTION}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "CompanyName" "${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "LegalTrademarks" "${APPNAME} is a trademark of ${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "LegalCopyright" "${DEVELOPER} | ${COMPANYNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "FileDescription" "${APPNAME}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "FileVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
  VIAddVersionKey /LANG=${LANG_ENGLISH} "ProductVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}.${BUILDNUMBER}"

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
  rmDir /r "$INSTDIR"

  #Delete installation folder from registry if available - this will only happen if it is empty
  DeleteRegKey /ifempty HKCU "Software\${APPNAME}"

  # Remove uninstaller information from the registry
  DeleteRegKey HKLM "Software\Microstft\Windows\CurrentVersion\Uninstall\${APPNAME}"
SectionEnd

;--------------------------------