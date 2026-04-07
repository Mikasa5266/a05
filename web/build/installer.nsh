!macro customInstall
  WriteRegStr HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "Interview AI" "$\"$INSTDIR\\${APP_EXECUTABLE_FILENAME}$\" --autostart"
!macroend

!macro customUnInstall
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Run" "Interview AI"
!macroend
