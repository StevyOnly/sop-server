; SOP Hub 服务安装包 - NSIS + NSSM 方案
!define APP_NAME    "SOP"
!define SERVICE     "sopsvc"
!define DISPLAY_NAME "SOP 后端服务"
!define VERSION     "1.0.0"
!define INSTDIR_DEF  "C:\Program Files\SOP"

Name "${APP_NAME}"
OutFile "SOP-Setup-${VERSION}.exe"
InstallDir "${INSTDIR_DEF}"
InstallDirRegKey HKLM "Software\SOP" "InstallLocation"

RequestExecutionLevel admin          ; 注册服务必须管理员
Unicode true
ShowInstDetails show                ; 显示安装细节，便于观察 nsExec 输出

; Pages: directory -> details
Page directory
Page instfiles
UninstPage uninstConfirm
UninstPage instfiles

!include "LogicLib.nsh"

Section "Install"
  ; Static license notice
  MessageBox MB_OK|MB_ICONINFORMATION "默认许可，直接下一步"

  SetOutPath "$INSTDIR"

  ; ---------- 先停掉可能正在运行的旧服务，避免文件被占用 ----------
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" stop ${SERVICE}'
  Sleep 1500
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" remove ${SERVICE} confirm'
  Sleep 1000

  ; Copy program & config (paths relative to install\)
  File "sop.exe"
  File "config.yaml"
  File "nssm.exe"
  ; Pre-create upload dir (sop uses relative path uploads/)
  CreateDirectory "$INSTDIR\uploads"

  ; ---------- 用 NSSM 安装服务 ----------
  ; 1) 注册服务（服务本体 = nssm.exe，应用 = sop.exe）
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" install ${SERVICE} "$INSTDIR\sop.exe"'
  Pop $0
  ${If} $0 != 0
    MessageBox MB_OK|MB_ICONSTOP "服务注册失败 (nssm install 返回 $0)"
    Abort
  ${EndIf}

  ; 2) 关键：工作目录 = 安装目录（否则 config.yaml/logs/uploads 都跑到 System32）
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} AppDirectory "$INSTDIR"'
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} AppStdout "$INSTDIR\logs\service.log"'
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} AppStderr "$INSTDIR\logs\service.err.log"'
  CreateDirectory "$INSTDIR\logs"

  ; 3) 开机自启 + 崩溃自动重启（指数退避）
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} Start SERVICE_AUTO_START'
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} AppExit Default Restart'
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" set ${SERVICE} AppExit 0 Restart'

  ; 4) 启动服务
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" start ${SERVICE}'
  ; 等待服务进入运行状态后再继续（最多 ~20s），避免安装完成但服务未就绪
  ; 通过 findstr 匹配 sc query 输出中的运行状态（兼容中英文系统）
  StrCpy $3 0
  ${Do}
    Sleep 1000
    nsExec::ExecToStack 'cmd /c sc query ${SERVICE} | findstr /i "RUNNING 运行"'
    Pop $1
    Pop $2
    StrLen $4 $2
    ${If} $4 > 0
      ${Break}
    ${EndIf}
    IntOp $3 $3 + 1
  ${LoopUntil} $3 >= 20
  ClearErrors

  ; 记录安装目录，供卸载与新装判断
  WriteRegStr HKLM "Software\SOP" "InstallLocation" "$INSTDIR"
  WriteUninstaller "$INSTDIR\uninstall.exe"

  ; ---------- 前端静态资源部署（sop_web 构建产物） ----------
  ; 每次安装都覆盖拷贝到 $INSTDIR\sop_web，保证升级后前端始终为最新构建
  SetOutPath "$INSTDIR\sop_web"
  File /r "sop_web\*.*"
SectionEnd

Section "Uninstall"
  ; 从注册表恢复真正的安装路径（防止控制面板卸载时 $INSTDIR 丢失）
  ReadRegStr $INSTDIR HKLM "Software\SOP" "InstallLocation"
  StrCmp $INSTDIR "" 0 +2
    StrCpy $INSTDIR "${INSTDIR_DEF}"

  ; 先停止并删除服务
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" stop ${SERVICE}'
  Sleep 2000
  nsExec::ExecToLog '"$INSTDIR\nssm.exe" remove ${SERVICE} confirm'
  ; 删除文件
  Delete "$INSTDIR\uninstall.exe"
  RMDir /r "$INSTDIR\sop_web"
  RMDir /r "$INSTDIR"
  DeleteRegKey HKLM "Software\SOP"
SectionEnd
