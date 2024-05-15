!include "MUI2.nsh"

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

!define NAME "Raiden"
!define APPFILE "raiden.exe"

Name "${NAME}"
Outfile "raiden-windows-${ARCH}-setup.exe"
InstallDir $WINDIR
ShowInstDetails show
RequestExecutionLevel admin

Section
  SetOutPath $INSTDIR
  File raiden.exe

  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Raiden" "DisplayName" "Raiden"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Raiden" "Publisher" "Refactory"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Raiden" "UninstallString" "$\"$INSTDIR\raiden-uninstall.exe$\""

  WriteUninstaller $INSTDIR\raiden-uninstall.exe
SectionEnd

Section "Uninstall"
  Delete $INSTDIR\raiden.exe
  Delete $INSTDIR\raiden-uninstall.exe

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Raiden"
SectionEnd
